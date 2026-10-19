// Package chat is direct messages and forums on top of end-to-end encryption. The server stores who is in a conversation,
// which wrapped keys each device holds, and ciphertext envelopes. It cannot read a message.
package chat

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/authz"
	"hub/internal/crypto"
	"hub/internal/e2ee"
	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/realtime"
	"hub/internal/tenant"
)

type Service struct{ A *app.App }

func New(a *app.App) *Service { return &Service{A: a} }

type KeyWrap struct {
	Epoch   int       `json:"epoch"`
	Device  uuid.UUID `json:"device"`
	Wrapped string    `json:"wrapped_key"`
}

type CreateIn struct {
	Kind        string      `json:"kind"`
	Name        string      `json:"name"`
	About       string      `json:"about"`
	Color       string      `json:"color"`
	Privacy     string      `json:"privacy"`
	Members     []uuid.UUID `json:"members"` // person ids; the creator is added automatically
	Audience    string      `json:"audience,omitempty"`
	AudienceRef *uuid.UUID  `json:"audience_ref,omitempty"`
	Keys        []KeyWrap   `json:"keys"`
}

// kindRule says who may create each kind of container and what shape it takes.
type kindRule struct {
	module, right string
	selfOnly      bool // only the creator (a personal notes vault)
	pair          bool // exactly one other person; created once per pair
	named         bool
	oneOnly       bool // one per creator (notes vault)
}

var kindRules = map[string]kindRule{
	"dm":      {module: "chat", right: "Send private messages", pair: true},
	"forum":   {module: "forums", right: "Create forum", named: true},
	"wall":    {module: "wall", right: "Moderate posts", named: true},
	"note":    {module: "home", right: "Customize home", selfOnly: true, oneOnly: true},
	"ticket":  {module: "tickets", right: "Raise"},
	"project": {module: "projects", right: "Create project", named: true},
	"event":   {module: "calendar", right: "Add events"},
	"folder":  {module: "files", right: "Upload", named: true},
}

// MaxAudience bounds how many people one end-to-end encrypted audience can hold: every member's devices each need a wrapped
// key, so the cost of adding or removing someone grows with the audience. Bigger organizations use one wall per branch or
// department (the same size limit is why group chats in other end-to-end encrypted systems are capped).
const MaxAudience = 5000

type Member struct {
	Person   uuid.UUID `json:"person"`
	Account  uuid.UUID `json:"account"`
	Name     string    `json:"name"`
	Initials string    `json:"initials"`
	Color    string    `json:"color"`
	Role     string    `json:"role"`
}

type Conversation struct {
	ID          uuid.UUID  `json:"id"`
	Kind        string     `json:"kind"`
	Name        string     `json:"name,omitempty"`
	About       string     `json:"about,omitempty"`
	Color       string     `json:"color,omitempty"`
	Initials    string     `json:"initials,omitempty"`
	Privacy     string     `json:"privacy"`
	Epoch       int        `json:"epoch"`
	MyRole      string     `json:"my_role,omitempty"`
	FirstEpoch  int        `json:"first_epoch,omitempty"`
	LastMessage *uuid.UUID `json:"last_message,omitempty"`
	LastAt      *time.Time `json:"last_at,omitempty"`
	Unread      int        `json:"unread"`
	Members     []Member   `json:"members,omitempty"`
	MemberCount int        `json:"member_count"`
	CreatedAt   time.Time  `json:"created_at"`
	Audience    string     `json:"audience,omitempty"`     // walls: org | sector | branch | unit | department
	AudienceRef *uuid.UUID `json:"audience_ref,omitempty"` // the structure unit the audience covers
}

func convCtx(org, id uuid.UUID, col string) crypto.Context {
	return crypto.Context{Org: org.String(), Table: "conversations", Column: col, RowID: id.String()}
}

func b64(s string) ([]byte, error) { return base64.StdEncoding.DecodeString(s) }

// deviceSet returns the active devices of the given accounts: device id -> account id.
func deviceSet(ctx context.Context, tx pgx.Tx, accounts []uuid.UUID) (map[uuid.UUID]uuid.UUID, error) {
	rs, err := tx.Query(ctx, `select id, account_id from hub.devices where account_id = any($1) and revoked_at is null`, accounts)
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	out := map[uuid.UUID]uuid.UUID{}
	for rs.Next() {
		var d, a uuid.UUID
		if err := rs.Scan(&d, &a); err != nil {
			return nil, err
		}
		out[d] = a
	}
	return out, rs.Err()
}

// checkCovers makes sure the wraps cover exactly the devices that must hold the key at this epoch. This is what stops a
// client (or someone who stole a session) from quietly leaving a member's device out, or adding a stranger's device.
func checkCovers(keys []KeyWrap, epoch int, devices map[uuid.UUID]uuid.UUID) error {
	got := map[uuid.UUID]bool{}
	for _, k := range keys {
		if k.Epoch != epoch {
			continue
		}
		if _, ok := devices[k.Device]; !ok {
			return httpx.Bad("A key was addressed to a device that is not part of this conversation.")
		}
		got[k.Device] = true
	}
	for d := range devices {
		if !got[d] {
			return httpx.Err(409, "keys_incomplete", "Every member's device needs a key.")
		}
	}
	return nil
}

func validateWraps(keys []KeyWrap) ([][]byte, error) {
	if len(keys) > 20000 {
		return nil, httpx.Bad("Too many keys in one request.")
	}
	raw := make([][]byte, len(keys))
	for i, k := range keys {
		b, err := b64(k.Wrapped)
		if err != nil || e2ee.ValidateWrapped(b) != nil || k.Epoch < 1 {
			return nil, httpx.Bad("A wrapped key is not valid.")
		}
		raw[i] = b
	}
	return raw, nil
}

func insertWraps(ctx context.Context, tx pgx.Tx, org, conv uuid.UUID, by *uuid.UUID, keys []KeyWrap, raw [][]byte) error {
	if len(keys) == 0 {
		return nil
	}
	b := &pgx.Batch{}
	for i, k := range keys {
		b.Queue(`insert into hub.conversation_keys (org_id, conversation_id, epoch, device_id, wrapped_key, wrapped_by) values ($1,$2,$3,$4,$5,$6) on conflict do nothing`, org, conv, k.Epoch, k.Device, raw[i], by)
	}
	br := tx.SendBatch(ctx, b)
	for range keys {
		if _, err := br.Exec(); err != nil {
			_ = br.Close()
			return err
		}
	}
	return br.Close()
}

// myDevice returns the caller's device that the request says it is acting as (needed to record who "wrapped" a key).
func (s *Service) myDevice(ctx context.Context, tx pgx.Tx, account, device uuid.UUID) error {
	var ok bool
	if err := tx.QueryRow(ctx, `select exists (select 1 from hub.devices where id=$1 and account_id=$2 and revoked_at is null)`, device, account).Scan(&ok); err != nil {
		return err
	}
	if !ok {
		return httpx.Bad("That device is not yours or was removed.")
	}
	return nil
}

func initialsOf(name string) string {
	out := []rune{}
	for _, w := range strings.Fields(name) {
		out = append(out, []rune(strings.ToUpper(w))[0])
		if len(out) == 2 {
			break
		}
	}
	return string(out)
}

func pairKey(a, b uuid.UUID) string {
	s := []string{a.String(), b.String()}
	sort.Strings(s)
	return s[0] + ":" + s[1]
}

// Create starts a direct message or a forum. A direct message between the same two people is created once: asking again
// returns the existing one.
func (s *Service) Create(ctx context.Context, p *tenant.Principal, in CreateIn) (*Conversation, bool, error) {
	rule, ok := kindRules[in.Kind]
	if !ok {
		return nil, false, httpx.Bad("Unknown kind of conversation.")
	}
	if in.Privacy == "" {
		in.Privacy = "invite"
	}
	if in.Privacy != "invite" && in.Privacy != "open" {
		return nil, false, httpx.Bad("Choose invite-only or open.")
	}
	in.Name = strings.TrimSpace(in.Name)
	if rule.named && (in.Name == "" || len(in.Name) > 80) || len(in.About) > 500 || len(in.Name) > 80 {
		return nil, false, httpx.Bad("Give it a name (up to 80 characters).")
	}
	if rule.pair && len(in.Members) != 1 {
		return nil, false, httpx.Bad("A direct message is between you and one person.")
	}
	if rule.selfOnly && len(in.Members) != 0 {
		return nil, false, httpx.Bad("This is just for you.")
	}
	if in.Audience != "" && in.Kind != "wall" && in.Kind != "folder" {
		return nil, false, httpx.Bad("Only a wall or a folder has an audience.")
	}
	if in.Color == "" {
		in.Color = "#7CC4FF"
	}
	if len(in.Color) != 7 || in.Color[0] != '#' {
		return nil, false, httpx.Bad("Choose a colour like #7CC4FF.")
	}
	raw, err := validateWraps(in.Keys)
	if err != nil {
		return nil, false, err
	}
	var out *Conversation
	existed := false
	err = s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, rule.module, rule.right); err != nil {
			return err
		}
		people := append([]uuid.UUID{p.Person}, in.Members...)
		if in.Audience != "" {
			ids, err := s.audienceMembers(ctx, tx, in.Audience, in.AudienceRef)
			if err != nil {
				return err
			}
			people = append([]uuid.UUID{p.Person}, ids...)
		}
		members, err := s.resolve(ctx, tx, people)
		if err != nil {
			return err
		}
		// for an audience, people without an account yet are simply not members until they join
		if len(members) != len(dedupe(people)) && in.Audience == "" {
			return httpx.Bad("Everyone you add needs an active account in this organization.")
		}
		var pk *string
		if rule.oneOnly {
			k := in.Kind + ":" + p.Account.String()
			pk = &k
			var id uuid.UUID
			if err := tx.QueryRow(ctx, `select id from hub.conversations where pair_key=$1`, k).Scan(&id); err == nil {
				existed = true
				out, err = s.get(ctx, tx, p, id)
				return err
			}
		}
		if rule.pair {
			other := members[1]
			if other.Account == p.Account {
				return httpx.Bad("You cannot message yourself here.")
			}
			k := pairKey(p.Account, other.Account)
			pk = &k
			var id uuid.UUID
			if err := tx.QueryRow(ctx, `select id from hub.conversations where pair_key=$1`, k).Scan(&id); err == nil {
				existed = true
				out, err = s.get(ctx, tx, p, id)
				return err
			}
		}
		accts := make([]uuid.UUID, len(members))
		for i, m := range members {
			accts[i] = m.Account
		}
		devs, err := deviceSet(ctx, tx, accts)
		if err != nil {
			return err
		}
		if len(devs) == 0 {
			return httpx.Err(409, "no_devices", "Set up secure messaging on this device first.")
		}
		if err := checkCovers(in.Keys, 1, devs); err != nil {
			return err
		}
		dek, _, err := s.A.OrgDEK(ctx, tx, p.Org)
		if err != nil {
			return err
		}
		id := app.NewID()
		nameEnc, _ := crypto.Field(dek, convCtx(p.Org, id, "name_enc"), []byte(in.Name))
		var aboutEnc []byte
		if in.About != "" {
			aboutEnc, _ = crypto.Field(dek, convCtx(p.Org, id, "about_enc"), []byte(in.About))
		}
		var aud *string
		if in.Audience != "" {
			aud = &in.Audience
		}
		if _, err := tx.Exec(ctx, `insert into hub.conversations (id, org_id, kind, name_enc, about_enc, color, initials, privacy, created_by, pair_key, audience, audience_ref) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
			id, p.Org, in.Kind, nameEnc, aboutEnc, in.Color, initialsOf(in.Name), in.Privacy, p.Account, pk, aud, in.AudienceRef); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "conversations", "id=$1", id); err != nil {
			return err
		}
		b := &pgx.Batch{}
		for i, m := range members {
			role := "member"
			if i == 0 {
				role = "owner"
			}
			b.Queue(`insert into hub.conversation_members (org_id, conversation_id, account_id, role) values ($1,$2,$3,$4)`, p.Org, id, m.Account, role)
		}
		br := tx.SendBatch(ctx, b)
		for range members {
			if _, err := br.Exec(); err != nil {
				_ = br.Close()
				return err
			}
		}
		if err := br.Close(); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "conversation_members", "conversation_id=$1", id); err != nil {
			return err
		}
		if err := insertWraps(ctx, tx, p.Org, id, nil, in.Keys, raw); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `insert into hub.conversation_stats (org_id, conversation_id) values ($1,$2)`, p.Org, id); err != nil {
			return err
		}
		if in.Kind == "forum" || in.Kind == "wall" {
			if err := s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Created "+in.Kind+" "+in.Name, "Forums", p.IP, map[string]any{"conversation": id}); err != nil {
				return err
			}
		}
		out, err = s.get(ctx, tx, p, id)
		return err
	})
	if err != nil {
		return nil, false, err
	}
	if !existed {
		for _, m := range out.Members {
			if m.Account != p.Account {
				d, _ := json.Marshal(map[string]any{"conversation": out.ID, "kind": out.Kind})
				s.A.Publish(ctx, realtime.Event{Type: "conversation.new", Channel: realtime.AcctChannel(m.Account), Org: p.Org.String(), Actor: p.Person.String(), Data: d})
			}
		}
	}
	return out, existed, nil
}

func dedupe(ids []uuid.UUID) []uuid.UUID {
	seen := map[uuid.UUID]bool{}
	out := []uuid.UUID{}
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	return out
}

// resolve turns person ids into active members with accounts, in the order given.
func (s *Service) resolve(ctx context.Context, tx pgx.Tx, people []uuid.UUID) ([]Member, error) {
	people = dedupe(people)
	rs, err := tx.Query(ctx, `select id, account_id, first_name || ' ' || last_name, color from hub.people where id = any($1) and status='active' and account_id is not null`, people)
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	by := map[uuid.UUID]Member{}
	for rs.Next() {
		var m Member
		if err := rs.Scan(&m.Person, &m.Account, &m.Name, &m.Color); err != nil {
			return nil, err
		}
		m.Initials = initialsOf(m.Name)
		by[m.Person] = m
	}
	out := []Member{}
	for _, id := range people {
		if m, ok := by[id]; ok {
			out = append(out, m)
		}
	}
	return out, rs.Err()
}

func (s *Service) members(ctx context.Context, tx pgx.Tx, id uuid.UUID) ([]Member, error) {
	rs, err := tx.Query(ctx, `select p.id, p.account_id, p.first_name || ' ' || p.last_name, p.color, m.role
		from hub.conversation_members m join hub.people p on p.account_id = m.account_id and p.org_id = m.org_id
		where m.conversation_id=$1 and m.left_at is null order by m.joined_at, p.id limit 500`, id)
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	out := []Member{}
	for rs.Next() {
		var m Member
		if err := rs.Scan(&m.Person, &m.Account, &m.Name, &m.Color, &m.Role); err != nil {
			return nil, err
		}
		m.Initials = initialsOf(m.Name)
		out = append(out, m)
	}
	return out, rs.Err()
}

const convSelect = `
	select c.id, c.kind, c.name_enc, c.about_enc, coalesce(c.color,''), coalesce(c.initials,''), c.privacy, c.epoch, c.created_at,
	       coalesce(c.audience,''), c.audience_ref, coalesce(me.role,''), coalesce(me.first_epoch,0), st.last_message, st.last_at,
	       (select count(*) from hub.conversation_members x where x.conversation_id = c.id and x.left_at is null),
	       case when me.account_id is null then 0 else (select count(*) from (select 1 from hub.messages mm where mm.org_id = c.org_id and mm.conversation_id = c.id and mm.sender_account <> $1
	            and mm.deleted_at is null and mm.id > coalesce(rc.last_read, '00000000-0000-0000-0000-000000000000'::uuid) limit 100) z) end
	from hub.conversations c
	left join hub.conversation_members me on me.conversation_id = c.id and me.account_id = $1 and me.left_at is null
	left join hub.conversation_stats st on st.conversation_id = c.id
	left join hub.read_cursors rc on rc.conversation_id = c.id and rc.account_id = $1`

func (s *Service) scanConv(ctx context.Context, tx pgx.Tx, org uuid.UUID, dek []byte, rs pgx.Rows) ([]*Conversation, error) {
	out := []*Conversation{}
	for rs.Next() {
		c := &Conversation{}
		var name, about []byte
		if err := rs.Scan(&c.ID, &c.Kind, &name, &about, &c.Color, &c.Initials, &c.Privacy, &c.Epoch, &c.CreatedAt, &c.Audience, &c.AudienceRef, &c.MyRole, &c.FirstEpoch, &c.LastMessage, &c.LastAt, &c.MemberCount, &c.Unread); err != nil {
			return nil, err
		}
		if len(name) > 0 {
			if b, err := crypto.FieldOpen(dek, convCtx(org, c.ID, "name_enc"), name); err == nil {
				c.Name = string(b)
			}
		}
		if len(about) > 0 {
			if b, err := crypto.FieldOpen(dek, convCtx(org, c.ID, "about_enc"), about); err == nil {
				c.About = string(b)
			}
		}
		out = append(out, c)
	}
	return out, rs.Err()
}

func (s *Service) get(ctx context.Context, tx pgx.Tx, p *tenant.Principal, id uuid.UUID) (*Conversation, error) {
	dek, _, err := s.A.OrgDEK(ctx, tx, p.Org)
	if err != nil {
		return nil, err
	}
	rs, err := tx.Query(ctx, convSelect+` where c.id = $2`, p.Account, id)
	if err != nil {
		return nil, err
	}
	list, err := s.scanConv(ctx, tx, p.Org, dek, rs)
	rs.Close()
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, httpx.ErrNotFound
	}
	c := list[0]
	// anyone may see an open forum's description; everything else needs membership
	if c.MyRole == "" && !(c.Kind == "forum" && c.Privacy == "open") {
		return nil, httpx.ErrNotFound
	}
	if c.MyRole != "" {
		if c.Members, err = s.members(ctx, tx, id); err != nil {
			return nil, err
		}
	}
	return c, nil
}

func (s *Service) Get(ctx context.Context, p *tenant.Principal, id uuid.UUID) (*Conversation, error) {
	var out *Conversation
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		var err error
		out, err = s.get(ctx, tx, p, id)
		return err
	})
	return out, err
}

// List returns the caller's conversations, most recent first. withMembers adds the participants of direct messages so the
// app can show the other person's name without another request.
func (s *Service) List(ctx context.Context, p *tenant.Principal, kind string, open bool, limit int, before *time.Time) ([]*Conversation, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	out := []*Conversation{}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		dek, _, err := s.A.OrgDEK(ctx, tx, p.Org)
		if err != nil {
			return err
		}
		where := "me.account_id is not null"
		args := []any{p.Account}
		if open {
			if err := authz.Require(ctx, tx, p.Person, "forums", "View"); err != nil {
				return err
			}
			where = "me.account_id is null and c.kind='forum' and c.privacy='open'"
		}
		if kind != "" {
			args = append(args, kind)
			where += fmt.Sprintf(" and c.kind = $%d", len(args))
		}
		if before != nil {
			args = append(args, *before)
			where += fmt.Sprintf(" and coalesce(st.last_at, c.created_at) < $%d", len(args))
		}
		rs, err := tx.Query(ctx, convSelect+" where "+where+fmt.Sprintf(" order by coalesce(st.last_at, c.created_at) desc, c.id desc limit %d", limit), args...)
		if err != nil {
			return err
		}
		out, err = s.scanConv(ctx, tx, p.Org, dek, rs)
		rs.Close()
		if err != nil {
			return err
		}
		for _, c := range out {
			if c.Kind == "dm" {
				if c.Members, err = s.members(ctx, tx, c.ID); err != nil {
					return err
				}
			}
		}
		return nil
	})
	return out, err
}

// Join adds the caller to an open forum. Others' devices then supply the key (see MissingKeys).
func (s *Service) Join(ctx context.Context, p *tenant.Principal, id uuid.UUID) error {
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "forums", "View"); err != nil {
			return err
		}
		var epoch int
		var priv, kind string
		if err := tx.QueryRow(ctx, `select epoch, privacy, kind from hub.conversations where id=$1`, id).Scan(&epoch, &priv, &kind); errors.Is(err, pgx.ErrNoRows) {
			return httpx.ErrNotFound
		} else if err != nil {
			return err
		}
		if kind != "forum" || priv != "open" {
			return httpx.ErrNotFound
		}
		if _, err := tx.Exec(ctx, `insert into hub.conversation_members (org_id, conversation_id, account_id, first_epoch) values ($1,$2,$3,$4)
			on conflict (org_id, conversation_id, account_id) do update set left_at=null, first_epoch=excluded.first_epoch, joined_at=now(), `+integrity.BumpOn("conversation_members")+` where hub.conversation_members.left_at is not null`, p.Org, id, p.Account, epoch); err != nil {
			return err
		}
		return s.A.Sign.Seal(ctx, tx, "conversation_members", "conversation_id=$1 and account_id=$2", id, p.Account)
	})
	if err != nil {
		return err
	}
	d, _ := json.Marshal(map[string]any{"person": p.Person, "account": p.Account})
	s.A.Publish(ctx, realtime.Event{Type: "member.joined", Channel: realtime.ConvChannel(id), Org: p.Org.String(), Actor: p.Person.String(), Data: d})
	return nil
}

// audienceMembers lists the people an audience covers. Only people with an account can hold keys, so only they are returned.
func (s *Service) audienceMembers(ctx context.Context, tx pgx.Tx, audience string, ref *uuid.UUID) ([]uuid.UUID, error) {
	col := map[string]string{"sector": "sector_id", "branch": "branch_id", "unit": "unit_id", "department": "department_id"}[audience]
	q := `select id from hub.people where status='active' and account_id is not null`
	var args []any
	switch {
	case audience == "org":
	case col != "" && ref != nil:
		q += ` and ` + col + ` = $1`
		args = append(args, *ref)
	default:
		return nil, httpx.Bad("Choose who the wall is for.")
	}
	rs, err := tx.Query(ctx, q+fmt.Sprintf(" order by id limit %d", MaxAudience+1), args...)
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	out := []uuid.UUID{}
	for rs.Next() {
		var id uuid.UUID
		if err := rs.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	if len(out) > MaxAudience {
		return nil, httpx.Err(413, "audience_too_large", fmt.Sprintf("An encrypted wall can hold up to %d people. Create one per branch or department instead.", MaxAudience))
	}
	return out, rs.Err()
}

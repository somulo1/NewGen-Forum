package chat

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/authz"
	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/realtime"
	"hub/internal/tenant"
)

// lockConv loads the conversation the caller belongs to and locks it so key changes happen one at a time.
func (s *Service) lockConv(ctx context.Context, tx pgx.Tx, p *tenant.Principal, id uuid.UUID) (kind, privacy, role string, epoch int, err error) {
	err = tx.QueryRow(ctx, `select c.kind, c.privacy, coalesce(m.role,''), c.epoch from hub.conversations c
		left join hub.conversation_members m on m.conversation_id=c.id and m.account_id=$2 and m.left_at is null
		where c.id=$1 for update of c`, id, p.Account).Scan(&kind, &privacy, &role, &epoch)
	if errors.Is(err, pgx.ErrNoRows) || err == nil && role == "" {
		return "", "", "", 0, httpx.ErrNotFound
	}
	return
}

// AddMembers invites people. keys must wrap the CURRENT conversation key for every device of every new member.
func (s *Service) AddMembers(ctx context.Context, p *tenant.Principal, id uuid.UUID, people []uuid.UUID, keys []KeyWrap, by uuid.UUID) ([]Member, error) {
	raw, err := validateWraps(keys)
	if err != nil {
		return nil, err
	}
	var added []Member
	err = s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		kind, _, role, epoch, err := s.lockConv(ctx, tx, p, id)
		if err != nil {
			return err
		}
		if kind == "dm" || kind == "note" || kind == "wall" {
			return httpx.Bad("People cannot be added to this kind of conversation by hand.")
		}
		if role != "owner" && role != "admin" {
			if err := authz.Require(ctx, tx, p.Person, "forums", "Invite members"); err != nil {
				return err
			}
		}
		if err := s.myDevice(ctx, tx, p.Account, by); err != nil {
			return err
		}
		ms, err := s.resolve(ctx, tx, people)
		if err != nil {
			return err
		}
		if len(ms) != len(dedupe(people)) || len(ms) == 0 {
			return httpx.Bad("Everyone you add needs an active account in this organization.")
		}
		var accts []uuid.UUID
		for _, m := range ms {
			accts = append(accts, m.Account)
		}
		devs, err := deviceSet(ctx, tx, accts)
		if err != nil {
			return err
		}
		if err := checkCovers(keys, epoch, devs); err != nil {
			return err
		}
		for _, m := range ms {
			if _, err := tx.Exec(ctx, `insert into hub.conversation_members (org_id, conversation_id, account_id, first_epoch) values ($1,$2,$3,$4)
				on conflict (org_id, conversation_id, account_id) do update set left_at=null, first_epoch=excluded.first_epoch, joined_at=now(), `+integrity.BumpOn("conversation_members")+` where hub.conversation_members.left_at is not null`, p.Org, id, m.Account, epoch); err != nil {
				return err
			}
		}
		if err := s.A.Sign.Seal(ctx, tx, "conversation_members", "conversation_id=$1", id); err != nil {
			return err
		}
		if err := insertWraps(ctx, tx, p.Org, id, &by, keys, raw); err != nil {
			return err
		}
		added = ms
		return nil
	})
	if err != nil {
		return nil, err
	}
	for _, m := range added {
		d, _ := json.Marshal(map[string]any{"conversation": id})
		s.A.Publish(ctx, realtime.Event{Type: "conversation.new", Channel: realtime.AcctChannel(m.Account), Org: p.Org.String(), Actor: p.Person.String(), Data: d})
	}
	d, _ := json.Marshal(map[string]any{"people": people})
	s.A.Publish(ctx, realtime.Event{Type: "member.added", Channel: realtime.ConvChannel(id), Org: p.Org.String(), Actor: p.Person.String(), Data: d})
	return added, nil
}

// RemoveMember takes someone out (or lets them leave). Everyone who stays gets a NEW conversation key in the same step,
// so the person who left cannot read anything sent after this moment. keys must wrap the next-epoch key for every
// remaining device.
func (s *Service) RemoveMember(ctx context.Context, p *tenant.Principal, id, person uuid.UUID, keys []KeyWrap, by uuid.UUID) error {
	raw, err := validateWraps(keys)
	if err != nil {
		return err
	}
	var leaving uuid.UUID
	err = s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		kind, _, role, epoch, err := s.lockConv(ctx, tx, p, id)
		if err != nil {
			return err
		}
		if kind == "dm" {
			return httpx.Bad("A direct message cannot be left. Mute it instead.")
		}
		var target uuid.UUID
		if err := tx.QueryRow(ctx, `select account_id from hub.people where id=$1 and account_id is not null`, person).Scan(&target); err != nil {
			return httpx.ErrNotFound
		}
		leaving = target
		if target != p.Account && role != "owner" && role != "admin" {
			if err := authz.Require(ctx, tx, p.Person, "forums", "Moderate"); err != nil {
				return err
			}
		}
		if err := s.myDevice(ctx, tx, p.Account, by); err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, `update hub.conversation_members set left_at=now(), `+integrity.Bump+` where conversation_id=$1 and account_id=$2 and left_at is null`, id, target)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return httpx.ErrNotFound
		}
		if err := s.A.Sign.Seal(ctx, tx, "conversation_members", "conversation_id=$1 and account_id=$2", id, target); err != nil {
			return err
		}
		rest, err := tx.Query(ctx, `select account_id from hub.conversation_members where conversation_id=$1 and left_at is null`, id)
		if err != nil {
			return err
		}
		var accts []uuid.UUID
		for rest.Next() {
			var a uuid.UUID
			if err := rest.Scan(&a); err != nil {
				rest.Close()
				return err
			}
			accts = append(accts, a)
		}
		rest.Close()
		if len(accts) > 0 {
			devs, err := deviceSet(ctx, tx, accts)
			if err != nil {
				return err
			}
			if err := checkCovers(keys, epoch+1, devs); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `update hub.conversations set epoch=epoch+1, `+integrity.Bump+` where id=$1`, id); err != nil {
				return err
			}
			if err := s.A.Sign.Seal(ctx, tx, "conversations", "id=$1", id); err != nil {
				return err
			}
			if err := insertWraps(ctx, tx, p.Org, id, &by, keys, raw); err != nil {
				return err
			}
		}
		return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Removed a member from a conversation", "Forums", p.IP, map[string]any{"conversation": id, "person": person})
	})
	if err != nil {
		return err
	}
	s.A.Hub.Detach(leaving, realtime.ConvChannel(id)) // other nodes: the removed account's socket checks membership on next subscribe
	d, _ := json.Marshal(map[string]any{"person": person})
	s.A.Publish(ctx, realtime.Event{Type: "member.removed", Channel: realtime.ConvChannel(id), Org: p.Org.String(), Actor: p.Person.String(), Data: d})
	s.A.Publish(ctx, realtime.Event{Type: "conversation.removed", Channel: realtime.AcctChannel(leaving), Org: p.Org.String(), Data: json.RawMessage(`{"conversation":"` + id.String() + `"}`)})
	return nil
}

// AddKeys lets a member's device hand the key to a device that lacks it (a colleague who just joined, or someone's new phone).
func (s *Service) AddKeys(ctx context.Context, p *tenant.Principal, id uuid.UUID, keys []KeyWrap, by uuid.UUID) (int, error) {
	raw, err := validateWraps(keys)
	if err != nil {
		return 0, err
	}
	n := 0
	err = s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		_, _, _, epoch, err := s.lockConv(ctx, tx, p, id)
		if err != nil {
			return err
		}
		if err := s.myDevice(ctx, tx, p.Account, by); err != nil {
			return err
		}
		rs, err := tx.Query(ctx, `select d.id, m.first_epoch from hub.devices d join hub.conversation_members m on m.account_id = d.account_id
			where m.conversation_id=$1 and m.left_at is null and d.revoked_at is null`, id)
		if err != nil {
			return err
		}
		first := map[uuid.UUID]int{}
		for rs.Next() {
			var d uuid.UUID
			var f int
			if err := rs.Scan(&d, &f); err != nil {
				rs.Close()
				return err
			}
			first[d] = f
		}
		rs.Close()
		for _, k := range keys {
			f, ok := first[k.Device]
			if !ok || k.Epoch > epoch || k.Epoch < f {
				return httpx.Bad("A key was addressed to a device that may not hold it.")
			}
		}
		if err := insertWraps(ctx, tx, p.Org, id, &by, keys, raw); err != nil {
			return err
		}
		n = len(keys)
		return nil
	})
	if err == nil {
		s.A.Publish(ctx, realtime.Event{Type: "keys.updated", Channel: realtime.ConvChannel(id), Org: p.Org.String()})
	}
	return n, err
}

type Missing struct {
	Device   uuid.UUID `json:"device"`
	Account  uuid.UUID `json:"account"`
	Identity string    `json:"identity_key"`
	Epoch    int       `json:"epoch"`
}

// MissingKeys lists member devices that have no key for the current epoch: work for any member's device to do.
func (s *Service) MissingKeys(ctx context.Context, p *tenant.Principal, id uuid.UUID) ([]Missing, error) {
	out := []Missing{}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		_, _, _, epoch, err := s.lockConvNoLock(ctx, tx, p, id)
		if err != nil {
			return err
		}
		rs, err := tx.Query(ctx, `select d.id, d.account_id, d.identity_key from hub.devices d
			join hub.conversation_members m on m.account_id = d.account_id and m.conversation_id=$1 and m.left_at is null
			where d.revoked_at is null and not exists (select 1 from hub.conversation_keys k where k.conversation_id=$1 and k.epoch=$2 and k.device_id = d.id)
			order by d.created_at limit 1000`, id, epoch)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var m Missing
			var ik []byte
			if err := rs.Scan(&m.Device, &m.Account, &ik); err != nil {
				return err
			}
			m.Identity, m.Epoch = base64.StdEncoding.EncodeToString(ik), epoch
			out = append(out, m)
		}
		return rs.Err()
	})
	return out, err
}

func (s *Service) lockConvNoLock(ctx context.Context, tx pgx.Tx, p *tenant.Principal, id uuid.UUID) (kind, privacy, role string, epoch int, err error) {
	err = tx.QueryRow(ctx, `select c.kind, c.privacy, coalesce(m.role,''), c.epoch from hub.conversations c
		left join hub.conversation_members m on m.conversation_id=c.id and m.account_id=$2 and m.left_at is null where c.id=$1`, id, p.Account).Scan(&kind, &privacy, &role, &epoch)
	if errors.Is(err, pgx.ErrNoRows) || err == nil && role == "" {
		return "", "", "", 0, httpx.ErrNotFound
	}
	return
}

type MyKey struct {
	Epoch   int       `json:"epoch"`
	Device  uuid.UUID `json:"device"`
	Wrapped string    `json:"wrapped_key"`
	By      *string   `json:"wrapped_by_signing_key,omitempty"`
}

// Keys returns the wrapped keys addressed to the caller's devices for every epoch they may read.
func (s *Service) Keys(ctx context.Context, p *tenant.Principal, id uuid.UUID) ([]MyKey, error) {
	out := []MyKey{}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if _, _, _, _, err := s.lockConvNoLock(ctx, tx, p, id); err != nil {
			return err
		}
		rs, err := tx.Query(ctx, `select k.epoch, k.device_id, k.wrapped_key from hub.conversation_keys k
			join hub.devices d on d.id = k.device_id and d.account_id=$2 and d.revoked_at is null
			join hub.conversation_members m on m.conversation_id = k.conversation_id and m.account_id = $2
			where k.conversation_id=$1 and k.epoch >= m.first_epoch order by k.epoch`, id, p.Account)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var k MyKey
			var w []byte
			if err := rs.Scan(&k.Epoch, &k.Device, &w); err != nil {
				return err
			}
			k.Wrapped = base64.StdEncoding.EncodeToString(w)
			out = append(out, k)
		}
		return rs.Err()
	})
	return out, err
}

func (s *Service) MarkRead(ctx context.Context, p *tenant.Principal, id, upTo uuid.UUID) error {
	return s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if _, _, _, _, err := s.lockConvNoLock(ctx, tx, p, id); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `insert into hub.read_cursors (org_id, conversation_id, account_id, last_read) values ($1,$2,$3,$4)
			on conflict (org_id, conversation_id, account_id) do update set last_read = greatest(hub.read_cursors.last_read, excluded.last_read), updated_at=now()`, p.Org, id, p.Account, upTo)
		return err
	})
}

var _ = time.Now

type ForumRules struct {
	PostMode string      `json:"post_mode"` // everyone | roles | list
	Roles    []string    `json:"roles"`     // role keys allowed to post when post_mode is "roles"
	Blocked  []uuid.UUID `json:"blocked"`
	Allowed  []uuid.UUID `json:"allowed"`
}

func (s *Service) Rules(ctx context.Context, p *tenant.Principal, id uuid.UUID) (ForumRules, error) {
	r := ForumRules{PostMode: "everyone", Roles: []string{}, Blocked: []uuid.UUID{}, Allowed: []uuid.UUID{}}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if _, _, _, _, err := s.lockConvNoLock(ctx, tx, p, id); err != nil {
			return err
		}
		err := tx.QueryRow(ctx, `select post_mode, roles, blocked, allowed from hub.forum_settings where conversation_id=$1`, id).Scan(&r.PostMode, &r.Roles, &r.Blocked, &r.Allowed)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	})
	return r, err
}

func (s *Service) SetRules(ctx context.Context, p *tenant.Principal, id uuid.UUID, r ForumRules) error {
	if r.PostMode != "everyone" && r.PostMode != "roles" && r.PostMode != "list" {
		return httpx.Bad("Choose everyone, certain roles, or a list of people.")
	}
	if r.Roles == nil {
		r.Roles = []string{}
	}
	if r.Blocked == nil {
		r.Blocked = []uuid.UUID{}
	}
	if r.Allowed == nil {
		r.Allowed = []uuid.UUID{}
	}
	return s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		kind, _, role, _, err := s.lockConvNoLock(ctx, tx, p, id)
		if err != nil {
			return err
		}
		if kind != "forum" {
			return httpx.Bad("Posting rules are for forums.")
		}
		if role != "owner" && role != "admin" {
			if err := authz.Require(ctx, tx, p.Person, "forums", "Moderate"); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(ctx, `insert into hub.forum_settings (org_id, conversation_id, post_mode, roles, blocked, allowed) values ($1,$2,$3,$4,$5,$6)
			on conflict (org_id, conversation_id) do update set post_mode=excluded.post_mode, roles=excluded.roles, blocked=excluded.blocked, allowed=excluded.allowed, `+integrity.BumpOn("forum_settings"), p.Org, id, r.PostMode, r.Roles, r.Blocked, r.Allowed); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "forum_settings", "conversation_id=$1", id); err != nil {
			return err
		}
		return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Changed who can post in a forum", "Forums", p.IP, r)
	})
}

// mayPost applies a forum's posting rules to a person.
func mayPost(ctx context.Context, tx pgx.Tx, conv, person uuid.UUID) error {
	var mode string
	var roles []string
	var blocked, allowed []uuid.UUID
	err := tx.QueryRow(ctx, `select post_mode, roles, blocked, allowed from hub.forum_settings where conversation_id=$1`, conv).Scan(&mode, &roles, &blocked, &allowed)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	deny := httpx.Err(403, "posting_limited", "You cannot post in this forum.")
	for _, b := range blocked {
		if b == person {
			return deny
		}
	}
	for _, a := range allowed {
		if a == person {
			return nil
		}
	}
	switch mode {
	case "list":
		return deny
	case "roles":
		var key string
		_ = tx.QueryRow(ctx, `select r.key from hub.person_roles pr join hub.roles r on r.id = pr.role_id where pr.person_id=$1`, person).Scan(&key)
		for _, r := range roles {
			if r == key {
				return nil
			}
		}
		return deny
	}
	return nil
}

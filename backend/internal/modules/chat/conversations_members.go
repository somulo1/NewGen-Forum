package chat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/authz"
	"hub/internal/crypto"
	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/realtime"
	"hub/internal/tenant"
)

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
	       coalesce(c.audience,''), c.audience_ref, coalesce(me.role,''), coalesce(me.first_epoch,0), st.last_message, st.last_at, st.changed_at,
	       coalesce((select cnt.members from hub.conversation_counts cnt where cnt.conversation_id = c.id), 0), -- kept count, not recounted
	       case when me.account_id is null then 0 else (select count(*) from (select 1 from hub.messages mm where mm.org_id = c.org_id and mm.conversation_id = c.id and mm.sender_account <> $1
	            and mm.deleted_at is null and mm.created_at > greatest(coalesce(cl.cleared_at, '-infinity'::timestamptz), case when c.kind='forum' then me.joined_at else '-infinity'::timestamptz end) and mm.id > coalesce(rc.last_read, '00000000-0000-0000-0000-000000000000'::uuid) limit 100) z) end
	from hub.conversations c
	left join hub.conversation_members me on me.conversation_id = c.id and me.account_id = $1 and me.left_at is null
	left join hub.conversation_stats st on st.conversation_id = c.id
	left join hub.read_cursors rc on rc.conversation_id = c.id and rc.account_id = $1
	left join hub.conversation_clears cl on cl.conversation_id = c.id and cl.account_id = $1`

func (s *Service) scanConv(ctx context.Context, tx pgx.Tx, org uuid.UUID, dek []byte, rs pgx.Rows) ([]*Conversation, error) {
	out := []*Conversation{}
	for rs.Next() {
		c := &Conversation{}
		var name, about []byte
		if err := rs.Scan(&c.ID, &c.Kind, &name, &about, &c.Color, &c.Initials, &c.Privacy, &c.Epoch, &c.CreatedAt, &c.Audience, &c.AudienceRef, &c.MyRole, &c.FirstEpoch, &c.LastMessage, &c.LastAt, &c.ChangedAt, &c.MemberCount, &c.Unread); err != nil {
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
		// a chat or forum I deleted stays out of my list until someone writes there again
		where := "me.account_id is not null and (cl.cleared_at is null or coalesce(st.last_at, c.created_at) > cl.cleared_at)"
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
		// the members of every private chat on the page in one query (not one per chat)
		var dms []uuid.UUID
		byID := map[uuid.UUID]*Conversation{}
		for _, c := range out {
			if c.Kind == "dm" {
				dms = append(dms, c.ID)
				byID[c.ID] = c
				c.Members = []Member{}
			}
		}
		if len(dms) == 0 {
			return nil
		}
		ms, err := tx.Query(ctx, `select m.conversation_id, p.id, p.account_id, p.first_name || ' ' || p.last_name, p.color, m.role
			from hub.conversation_members m join hub.people p on p.account_id = m.account_id and p.org_id = m.org_id
			where m.conversation_id = any($1) and m.left_at is null order by m.conversation_id, m.joined_at, p.id`, dms)
		if err != nil {
			return err
		}
		defer ms.Close()
		for ms.Next() {
			var cid uuid.UUID
			var m Member
			if err := ms.Scan(&cid, &m.Person, &m.Account, &m.Name, &m.Color, &m.Role); err != nil {
				return err
			}
			m.Initials = initialsOf(m.Name)
			byID[cid].Members = append(byID[cid].Members, m)
		}
		return ms.Err()
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

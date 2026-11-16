package chat

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/realtime"
	"hub/internal/tenant"
)

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

// SetMemberRole changes what someone can do in a shared note: "admin" can edit it, "member" can only read it.
func (s *Service) SetMemberRole(ctx context.Context, p *tenant.Principal, id, person uuid.UUID, role string) error {
	if role != "admin" && role != "member" {
		return httpx.Bad("Choose can edit or read only.")
	}
	return s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		kind, _, mine, _, err := s.lockConv(ctx, tx, p, id)
		if err != nil {
			return err
		}
		if kind != "sharednote" || mine != "owner" {
			return httpx.ErrForbidden
		}
		tag, err := tx.Exec(ctx, `update hub.conversation_members set role=$3, `+integrity.Bump+` where conversation_id=$1 and left_at is null and role <> 'owner' and account_id=(select account_id from hub.people where id=$2)`, id, person, role)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return httpx.ErrNotFound
		}
		return s.A.Sign.Seal(ctx, tx, "conversation_members", "conversation_id=$1", id)
	})
}

// tellKeyed sends "keys.shared" to the accounts whose devices were just given a key to a conversation.
func (s *Service) tellKeyed(ctx context.Context, org, conv uuid.UUID, keys []KeyWrap) {
	devs := make([]uuid.UUID, 0, len(keys))
	for _, k := range keys {
		devs = append(devs, k.Device)
	}
	var accts []uuid.UUID
	if err := s.A.DB.WithOrg(ctx, org, func(tx pgx.Tx) error {
		rs, err := tx.Query(ctx, `select distinct account_id from hub.devices where id = any($1)`, devs)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var a uuid.UUID
			if err := rs.Scan(&a); err != nil {
				return err
			}
			accts = append(accts, a)
		}
		return rs.Err()
	}); err != nil {
		return
	}
	d, _ := json.Marshal(map[string]any{"conversation": conv})
	for _, a := range accts {
		s.A.Publish(ctx, realtime.Event{Type: "keys.shared", Channel: realtime.AcctChannel(a), Org: org.String(), Data: d})
	}
}

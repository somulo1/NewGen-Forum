package chat

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/realtime"
	"hub/internal/tenant"
)

// Evict removes an account from every conversation of the organization (it was blocked, deactivated or deleted) and marks
// the conversations that still have people in them as needing a new key. The server cannot make that key, so one of the
// remaining members' devices does it the next time it asks what needs doing (see Todo).
func Evict(ctx context.Context, tx pgx.Tx, sign *integrity.Signer, org, account uuid.UUID) error {
	rs, err := tx.Query(ctx, `update hub.conversation_members set left_at=now(), `+integrity.Bump+` where org_id=$1 and account_id=$2 and left_at is null returning conversation_id`, org, account)
	if err != nil {
		return err
	}
	var convs []uuid.UUID
	for rs.Next() {
		var c uuid.UUID
		if err := rs.Scan(&c); err != nil {
			rs.Close()
			return err
		}
		convs = append(convs, c)
	}
	rs.Close()
	if len(convs) == 0 {
		return nil
	}
	if err := sign.Seal(ctx, tx, "conversation_members", "org_id=$1 and account_id=$2", org, account); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `update hub.conversations c set rotate_needed = true, `+integrity.Bump+`
		where c.id = any($1) and c.kind <> 'dm' and exists (select 1 from hub.conversation_members m where m.conversation_id = c.id and m.left_at is null)`, convs); err != nil {
		return err
	}
	return sign.Seal(ctx, tx, "conversations", "id = any($1)", convs)
}

// SyncAudience puts a person into the wall audiences that cover them (their branch, department, the whole organization).
// Their devices still need a key; see MissingKeys.
func SyncAudience(ctx context.Context, tx pgx.Tx, sign *integrity.Signer, org, account uuid.UUID) error {
	rs, err := tx.Query(ctx, `insert into hub.conversation_members (org_id, conversation_id, account_id, first_epoch)
		select c.org_id, c.id, p.account_id, c.epoch from hub.conversations c, hub.people p
		where p.account_id = $1 and p.status='active' and c.audience is not null and (
		   c.audience = 'org' or (c.audience='sector' and c.audience_ref = p.sector_id) or (c.audience='branch' and c.audience_ref = p.branch_id)
		   or (c.audience='unit' and c.audience_ref = p.unit_id) or (c.audience='department' and c.audience_ref = p.department_id))
		on conflict (org_id, conversation_id, account_id) do nothing returning conversation_id`, account)
	if err != nil {
		return err
	}
	n := 0
	for rs.Next() {
		n++
	}
	rs.Close()
	// leave audiences that no longer cover them (they changed branch or department); those walls then need a new key
	gone, err := tx.Query(ctx, `update hub.conversation_members m set left_at = now(), ver = m.ver + 1, mac = hub.pending_mac()
		from hub.conversations c, hub.people p
		where m.conversation_id = c.id and p.account_id = m.account_id and m.account_id = $1 and m.left_at is null and c.audience is not null
		  and not (c.audience = 'org' or (c.audience='sector' and c.audience_ref = p.sector_id) or (c.audience='branch' and c.audience_ref = p.branch_id)
		   or (c.audience='unit' and c.audience_ref = p.unit_id) or (c.audience='department' and c.audience_ref = p.department_id))
		returning c.id`, account)
	if err != nil {
		return err
	}
	var left []uuid.UUID
	for gone.Next() {
		var c uuid.UUID
		if err := gone.Scan(&c); err != nil {
			gone.Close()
			return err
		}
		left = append(left, c)
	}
	gone.Close()
	if n == 0 && len(left) == 0 {
		return nil
	}
	if err := sign.Seal(ctx, tx, "conversation_members", "account_id=$1", account); err != nil {
		return err
	}
	if len(left) > 0 {
		if _, err := tx.Exec(ctx, `update hub.conversations set rotate_needed=true, `+integrity.Bump+` where id = any($1)`, left); err != nil {
			return err
		}
		return sign.Seal(ctx, tx, "conversations", "id = any($1)", left)
	}
	return nil
}

type Todo struct {
	Conversation uuid.UUID `json:"conversation"`
	Kind         string    `json:"kind"`
	Missing      int       `json:"missing_devices"`
	Rotate       bool      `json:"rotate"`
}

// Todo is the work for this person's devices: conversations where colleagues' devices lack a key, or where someone left and
// a fresh key is due. A busy client runs it on connect and when it hears "member.joined" / "keys.updated".
func (s *Service) Todo(ctx context.Context, p *tenant.Principal) ([]Todo, error) {
	out := []Todo{}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		rs, err := tx.Query(ctx, `
			select c.id, c.kind, c.rotate_needed,
			  (select count(*) from hub.devices d join hub.conversation_members m2 on m2.account_id = d.account_id and m2.conversation_id = c.id and m2.left_at is null
			   where d.revoked_at is null and not exists (select 1 from hub.conversation_keys k where k.conversation_id = c.id and k.epoch = c.epoch and k.device_id = d.id))
			from hub.conversations c join hub.conversation_members me on me.conversation_id = c.id and me.account_id=$1 and me.left_at is null
			where c.rotate_needed or exists (select 1 from hub.devices d join hub.conversation_members m2 on m2.account_id = d.account_id and m2.conversation_id = c.id and m2.left_at is null
			   where d.revoked_at is null and not exists (select 1 from hub.conversation_keys k where k.conversation_id = c.id and k.epoch = c.epoch and k.device_id = d.id))
			limit 200`, p.Account)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var t Todo
			if err := rs.Scan(&t.Conversation, &t.Kind, &t.Rotate, &t.Missing); err != nil {
				return err
			}
			out = append(out, t)
		}
		return rs.Err()
	})
	return out, err
}

// Rotate starts a new key epoch for everyone who is in the conversation now (used after someone was evicted, or on request).
func (s *Service) Rotate(ctx context.Context, p *tenant.Principal, id uuid.UUID, keys []KeyWrap, by uuid.UUID) error {
	raw, err := validateWraps(keys)
	if err != nil {
		return err
	}
	err = s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		_, _, _, epoch, err := s.lockConv(ctx, tx, p, id)
		if err != nil {
			return err
		}
		if err := s.myDevice(ctx, tx, p.Account, by); err != nil {
			return err
		}
		rs, err := tx.Query(ctx, `select account_id from hub.conversation_members where conversation_id=$1 and left_at is null`, id)
		if err != nil {
			return err
		}
		var accts []uuid.UUID
		for rs.Next() {
			var a uuid.UUID
			if err := rs.Scan(&a); err != nil {
				rs.Close()
				return err
			}
			accts = append(accts, a)
		}
		rs.Close()
		devs, err := deviceSet(ctx, tx, accts)
		if err != nil {
			return err
		}
		if len(devs) == 0 {
			return httpx.Bad("Nobody in the conversation has a device.")
		}
		if err := checkCovers(keys, epoch+1, devs); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `update hub.conversations set epoch=epoch+1, rotate_needed=false, `+integrity.Bump+` where id=$1`, id); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "conversations", "id=$1", id); err != nil {
			return err
		}
		return insertWraps(ctx, tx, p.Org, id, &by, keys, raw)
	})
	if err == nil {
		d, _ := json.Marshal(map[string]any{"conversation": id})
		s.A.Publish(ctx, realtime.Event{Type: "keys.rotated", Channel: realtime.ConvChannel(id), Org: p.Org.String(), Data: d})
	}
	return err
}

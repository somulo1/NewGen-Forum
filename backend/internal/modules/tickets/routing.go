package tickets

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/notify"
	"hub/internal/tenant"
)

// The categories a place sets on its dashboard are the one source of who handles what: a ticket goes to whoever that place's
// category names (the one with the fewest open tickets). A business request or communication, which has no such list, goes to the
// heads of the place. Only people already on the ticket can be picked, since its contents are encrypted for them.

// routeCandidates are the people who should handle the ticket where it is now and who are on its conversation.
func (s *Service) routeCandidates(ctx context.Context, tx pgx.Tx, t *Ticket) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	cat, err := s.category(ctx, tx, t.ToType, t.ToRef, t.Category)
	if err != nil {
		return nil, err
	}
	q := `select p.id from hub.people p join hub.conversation_members m on m.account_id = p.account_id and m.org_id = p.org_id and m.left_at is null
		where m.conversation_id = $1 and p.status = 'active' and p.id <> $2 and p.id = any($3)`
	if cat != nil && len(cat.Members) > 0 {
		rs, err := tx.Query(ctx, q, t.ID, t.RaisedBy, cat.Members)
		if err != nil {
			return nil, err
		}
		for rs.Next() {
			var id uuid.UUID
			if err := rs.Scan(&id); err != nil {
				rs.Close()
				return nil, err
			}
			ids = append(ids, id)
		}
		rs.Close()
		if len(ids) > 0 || len(cat.Members) > 0 && t.Kind == "ticket" {
			return ids, rs.Err()
		}
	}
	if t.ToRef == nil {
		return ids, nil
	}
	// no one named for it: the heads of the place it is with
	rs, err := tx.Query(ctx, `select distinct p.id from hub.people p join hub.person_roles pr on pr.person_id = p.id join hub.roles r on r.id = pr.role_id
		join hub.conversation_members m on m.account_id = p.account_id and m.org_id = p.org_id and m.left_at is null and m.conversation_id = $2
		where p.status = 'active' and p.id <> $3 and $1 in (p.department_id, p.unit_id, p.branch_id, p.sector_id)
		  and coalesce(hub.chosen_head($1) = p.id, r.scope like 'Each %' or exists (select 1 from hub.unit_heads uh where uh.unit_id = $1 and uh.role_id = r.id))`, *t.ToRef, t.ID, t.RaisedBy)
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	for rs.Next() {
		var id uuid.UUID
		if err := rs.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rs.Err()
}

// autoAssign gives an unassigned open ticket to its place's handler (see routeCandidates). It reports whether it assigned someone.
func (s *Service) autoAssign(ctx context.Context, tx pgx.Tx, nb *notify.Batch, org uuid.UUID, t *Ticket) (bool, error) {
	if t.Handler != nil || t.Step >= 3 {
		return false, nil
	}
	ids, err := s.routeCandidates(ctx, tx, t)
	if err != nil || len(ids) == 0 {
		return false, err
	}
	who, err := s.pickHandler(ctx, tx, &Category{Members: ids})
	if err != nil || who == nil {
		return false, err
	}
	if _, err := tx.Exec(ctx, `update hub.tickets set handler=$2, handler_since=now(), step=greatest(step,1), updated_at=now(), `+integrity.Bump+` where conversation_id=$1 and handler is null`, t.ID, *who); err != nil {
		return false, err
	}
	if err := s.A.Sign.Seal(ctx, tx, "tickets", "conversation_id=$1", t.ID); err != nil {
		return false, err
	}
	if _, err := tx.Exec(ctx, `insert into hub.ticket_assignees (org_id, ticket_id, person_id) values ($1,$2,$3) on conflict do nothing`, org, t.ID, *who); err != nil {
		return false, err
	}
	if err := s.A.Sign.Seal(ctx, tx, "ticket_assignees", "ticket_id=$1", t.ID); err != nil {
		return false, err
	}
	var name string
	_ = tx.QueryRow(ctx, `select first_name || ' ' || last_name from hub.people where id=$1`, *who).Scan(&name)
	if err := s.event(ctx, tx, org, t.ID, nil, "event", "Assigned automatically to "+name+", who handles it where it is"); err != nil {
		return false, err
	}
	return true, nb.Add(ctx, tx, org, *who, "ticket", t.Ref+" was assigned to you.", "#7CC4FF", t.Ref)
}

// Route assigns an unassigned open ticket to whoever its place names for it. Someone on the ticket asks for it after adding those
// people to it (they hold its key); doing nothing when it already has someone.
func (s *Service) Route(ctx context.Context, p *tenant.Principal, id uuid.UUID) (*Ticket, error) {
	var out *Ticket
	var nb notify.Batch
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		cur, err := scanTicket(tx.QueryRow(ctx, ticketSelect+` where t.conversation_id=$1 for update of t`, id))
		if err != nil {
			return httpx.ErrNotFound
		}
		in, err := s.participant(ctx, tx, p, id)
		if err != nil {
			return err
		}
		if !in {
			return httpx.ErrNotFound
		}
		if _, err := s.autoAssign(ctx, tx, &nb, p.Org, cur); err != nil {
			return err
		}
		out, err = s.get(ctx, tx, p.Org, id)
		return err
	})
	if err != nil {
		return nil, err
	}
	nb.Send(ctx, s.A)
	s.changed(ctx, p, out)
	return out, nil
}

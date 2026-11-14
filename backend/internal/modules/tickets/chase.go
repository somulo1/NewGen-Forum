package tickets

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/notify"
	"hub/internal/tenant"
)

// Chasing an overdue ticket is not the usual escalation: the ticket stays where it is, with the same handler. The person who
// raised it, or a head of the dashboard it came from, writes a note and it goes straight to the heads of the dashboard it was
// sent to. Those heads are put on the ticket (so they can read it) and told at once.

// originOf is the dashboard a ticket came from: the place it was sent from (a sector, branch, unit or department dashboard),
// else the raiser's own place, the deepest of their sector, branch, unit and department (a person in a unit inside a department
// is from that unit; its heads, and the heads of every place above it, are heads of where it came from).
func originOf(ctx context.Context, tx pgx.Tx, ticket uuid.UUID) (*uuid.UUID, error) {
	var ref *uuid.UUID
	err := tx.QueryRow(ctx, `with t as (select t.from_ref, pe.sector_id, pe.branch_id, pe.unit_id, pe.department_id
			from hub.tickets t left join hub.people pe on pe.id = t.raised_by where t.conversation_id=$1),
		mine as (select x id from t, unnest(array[t.sector_id, t.branch_id, t.unit_id, t.department_id]) x where x is not null),
		depth as (select m.id, (with recursive up(id, parent_id, n) as (select u.id, u.parent_id, 0 from hub.org_units u where u.id = m.id
				union all select u.id, u.parent_id, up.n + 1 from hub.org_units u join up on u.id = up.parent_id) select max(n) from up) d from mine m)
		select coalesce((select from_ref from t), (select id from depth order by d desc nulls last limit 1))`, ticket).Scan(&ref)
	return ref, err
}

func overdue(t *Ticket) bool {
	return t.Step < 3 && (t.Late || (t.DueAt != nil && t.DueAt.Before(time.Now())))
}

// mayChase: the ticket is overdue, it went to a place with a head, and the caller raised it or heads the place it came from.
func (s *Service) mayChase(ctx context.Context, tx pgx.Tx, person uuid.UUID, t *Ticket) (bool, error) {
	if !overdue(t) || t.ToRef == nil || t.ToType == "" || t.ToType == "desk" {
		return false, nil
	}
	if t.RaisedBy == person {
		return true, nil
	}
	if t.OriginRef == nil {
		return false, nil
	}
	return headsPlace(ctx, tx, person, *t.OriginRef)
}

// placeHeads: the people who head exactly this place, the head of that unit, sector, branch or department, and not the heads
// of places above it. They are placed in it and hold either a role that heads each place of its kind ("Each unit") or the head
// role set for it in Organization structure.
const placeHeads = `select distinct pe.id from hub.org_units u join hub.people pe on pe.status = 'active'
		and u.id = case u.kind when 'sector' then pe.sector_id when 'branch' then pe.branch_id when 'unit' then pe.unit_id else pe.department_id end
	join hub.person_roles pr on pr.person_id = pe.id join hub.roles r on r.id = pr.role_id
	where u.id = $1 and coalesce(hub.chosen_head(u.id) = pe.id, lower(r.scope) like 'each ' || u.kind || '%' or exists (select 1 from hub.unit_heads uh where uh.unit_id = u.id and uh.role_id = r.id))`

func headsOf(ctx context.Context, tx pgx.Tx, unit uuid.UUID) ([]uuid.UUID, error) {
	rs, err := tx.Query(ctx, placeHeads+` limit 50`, unit)
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	var out []uuid.UUID
	for rs.Next() {
		var id uuid.UUID
		if err := rs.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rs.Err()
}

func headsPlace(ctx context.Context, tx pgx.Tx, person, unit uuid.UUID) (bool, error) {
	var ok bool
	err := tx.QueryRow(ctx, `select exists (`+placeHeads+` and pe.id = $2)`, unit, person).Scan(&ok)
	return ok, err
}

// Chase sends an overdue ticket, with a note, to the heads of the dashboard it was sent to.
func (s *Service) Chase(ctx context.Context, p *tenant.Principal, id uuid.UUID, note string) (*Ticket, error) {
	note, err := cleanText(note, maxComment, "The note")
	if err != nil {
		return nil, err
	}
	if note == "" {
		return nil, httpx.Bad("Write what the head should know.")
	}
	var out *Ticket
	var nb notify.Batch
	err = s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		t, err := s.get(ctx, tx, p.Org, id)
		if errors.Is(err, pgx.ErrNoRows) {
			return httpx.ErrNotFound
		} else if err != nil {
			return err
		}
		if t.OriginRef, err = originOf(ctx, tx, id); err != nil {
			return err
		}
		ok, err := s.mayChase(ctx, tx, p.Person, t)
		if err != nil {
			return err
		}
		if !ok {
			return httpx.Err(403, "not_allowed", "Only the person who raised an overdue ticket, or a head of the place it came from, can send it to the head.")
		}
		heads, err := headsOf(ctx, tx, *t.ToRef)
		if err != nil {
			return err
		}
		if len(heads) == 0 {
			return httpx.Err(409, "no_head", "The place it was sent to has no head yet. Ask an administrator to set one in Organization structure.")
		}
		// the heads join the ticket so they can read it
		var epoch int
		if err := tx.QueryRow(ctx, `select epoch from hub.conversations where id=$1`, id).Scan(&epoch); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `insert into hub.conversation_members (org_id, conversation_id, account_id, first_epoch)
				select $1, $2, pe.account_id, $3 from hub.people pe where pe.id = any($4) and pe.account_id is not null
				on conflict (org_id, conversation_id, account_id) do update set left_at=null, joined_at=now(), `+integrity.BumpOn("conversation_members")+`
				where hub.conversation_members.left_at is not null`, p.Org, id, epoch, heads); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "conversation_members", "conversation_id=$1", id); err != nil {
			return err
		}
		if err := s.event(ctx, tx, p.Org, id, p, "event", "Sent to the head of the place it is with, because it is overdue"); err != nil {
			return err
		}
		if err := s.addComment(ctx, tx, p.Org, id, &p.Person, "To the head: "+note); err != nil {
			return err
		}
		for _, h := range heads {
			if h == p.Person {
				continue
			}
			if err := nb.Add(ctx, tx, p.Org, h, "ticket", t.Ref+" is overdue. "+p.Name+" sent it to you: "+note, "#FF9B8A", t.Ref); err != nil {
				return err
			}
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

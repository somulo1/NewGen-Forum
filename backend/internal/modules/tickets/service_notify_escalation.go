package tickets

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/authz"
	"hub/internal/integrity"
	"hub/internal/notify"
)

// notifyEscalation tells whoever holds the escalation role (by role name or job title).
func (s *Service) notifyEscalation(ctx context.Context, tx pgx.Tx, nb *notify.Batch, org uuid.UUID, t *Ticket, cfg Config) error {
	rs, err := tx.Query(ctx, `select p.id from hub.people p left join hub.person_roles pr on pr.person_id = p.id left join hub.roles r on r.id = pr.role_id
		where p.status='active' and (lower(p.title) = lower($1) or lower(coalesce(r.name,'')) = lower($1)) limit 20`, cfg.Escalate)
	if err != nil {
		return err
	}
	var ids []uuid.UUID
	for rs.Next() {
		var id uuid.UUID
		if err := rs.Scan(&id); err != nil {
			rs.Close()
			return err
		}
		ids = append(ids, id)
	}
	rs.Close()
	for _, id := range ids {
		s.mailTicket(ctx, tx, nb, org, id, "escalated", t, nil, "", "It passed its target time.")
		if err := nb.Add(ctx, tx, org, id, "ticket", t.Ref+" was escalated to you.", "#FF9B8A", t.Ref); err != nil {
			return err
		}
	}
	return nil
}

// reportMissedTAT tells the heads of the place a ticket is with (people holding the place's head role and placed in it) that it
// passed its TAT. It returns how many were told.
func (s *Service) reportMissedTAT(ctx context.Context, tx pgx.Tx, nb *notify.Batch, org uuid.UUID, t *Ticket) (int, error) {
	if t.ToRef == nil || t.ToType == "desk" {
		return 0, nil
	}
	rs, err := tx.Query(ctx, `select distinct p.id from hub.unit_heads uh join hub.person_roles pr on pr.role_id = uh.role_id
		join hub.people p on p.id = pr.person_id and p.status = 'active'
		where uh.unit_id = $1 and (p.sector_id = $1 or p.branch_id = $1 or p.unit_id = $1 or p.department_id = $1) and coalesce(hub.chosen_head($1) = p.id, true)
		limit 20`, *t.ToRef)
	if err != nil {
		return 0, err
	}
	var ids []uuid.UUID
	for rs.Next() {
		var id uuid.UUID
		if err := rs.Scan(&id); err != nil {
			rs.Close()
			return 0, err
		}
		ids = append(ids, id)
	}
	rs.Close()
	for _, id := range ids {
		s.mailTicket(ctx, tx, nb, org, id, "escalated", t, nil, "", "It passed its target time.")
		if err := nb.Add(ctx, tx, org, id, "ticket", t.Ref+" passed its TAT and was reported to you.", "#FF9B8A", t.Ref); err != nil {
			return 0, err
		}
	}
	return len(ids), nil
}

// ---- timing sweep (run by the worker) ----

// Sweep marks tickets that ran past their TAT as late, escalates them, and warns about those nearing it. Safe to run on
// several nodes at once: each organization is swept under an advisory lock.
func (s *Service) Sweep(ctx context.Context, org uuid.UUID) (late, near int, err error) {
	var nb notify.Batch
	err = s.A.DB.WithOrg(ctx, org, func(tx pgx.Tx) error {
		var got bool
		if err := tx.QueryRow(ctx, `select pg_try_advisory_xact_lock(hashtextextended('tickets-sweep:'||$1::text, 0))`, org).Scan(&got); err != nil || !got {
			return err
		}
		cfg, err := s.config(ctx, tx)
		if err != nil {
			return err
		}
		if err := s.autoClose(ctx, tx, org, &nb); err != nil {
			return err
		}
		rs, err := tx.Query(ctx, `select t.conversation_id, t.ref, t.handler, t.created_at, t.due_at, t.late, t.escalated, t.hold, t.to_type, t.to_ref
			from hub.tickets t where t.step < 3 and t.due_at is not null and (not t.late and t.due_at < now() or (t.due_at - t.created_at) * ($1::numeric / 100) + t.created_at < now() and t.due_at >= now())
			order by t.due_at limit 1000`, cfg.NearPct) // a bounded batch per sweep: a backlog is worked through over the next sweeps
		if err != nil {
			return err
		}
		type row struct {
			id        uuid.UUID
			ref       string
			handler   *uuid.UUID
			created   time.Time
			due       time.Time
			late, esc bool
			hold      bool
			toType    string
			toRef     *uuid.UUID
		}
		var rowsSeen []row
		for rs.Next() {
			var r row
			if err := rs.Scan(&r.id, &r.ref, &r.handler, &r.created, &r.due, &r.late, &r.esc, &r.hold, &r.toType, &r.toRef); err != nil {
				rs.Close()
				return err
			}
			rowsSeen = append(rowsSeen, r)
		}
		rs.Close()
		// who was already told a ticket is nearing its deadline: one query for the whole batch, not one per ticket
		told := map[string]bool{}
		if refs := func() []string {
			var out []string
			for _, r := range rowsSeen {
				if r.handler != nil && !r.due.Before(time.Now()) {
					out = append(out, r.ref)
				}
			}
			return out
		}(); len(refs) > 0 && cfg.Notify.Nearing {
			nr, err := tx.Query(ctx, `select person_id, ref from hub.notifications where kind='ticket-near' and ref = any($1)`, refs)
			if err != nil {
				return err
			}
			for nr.Next() {
				var pid uuid.UUID
				var ref string
				if err := nr.Scan(&pid, &ref); err != nil {
					nr.Close()
					return err
				}
				told[pid.String()+"|"+ref] = true
			}
			nr.Close()
		}
		now := time.Now()
		for _, r := range rowsSeen {
			if r.hold && cfg.Pause.Onhold {
				continue
			}
			if r.due.Before(now) {
				if _, err := tx.Exec(ctx, `update hub.tickets set late=true, escalated=true, updated_at=now(), `+integrity.Bump+` where conversation_id=$1`, r.id); err != nil {
					return err
				}
				if err := s.A.Sign.Seal(ctx, tx, "tickets", "conversation_id=$1", r.id); err != nil {
					return err
				}
				// a missed TAT is reported to the heads of the place it is with (it stays there); without heads, to the escalation contact
				t := &Ticket{ID: r.id, Ref: r.ref, ToType: r.toType, ToRef: r.toRef}
				heads, err := s.reportMissedTAT(ctx, tx, &nb, org, t)
				if err != nil {
					return err
				}
				to := "the head of the place it is with"
				if heads == 0 {
					to = cfg.Escalate
					if err := s.notifyEscalation(ctx, tx, &nb, org, t, cfg); err != nil {
						return err
					}
				}
				if err := s.event(ctx, tx, org, r.id, nil, "event", "Passed its TAT: reported to "+to); err != nil {
					return err
				}
				if r.handler != nil && cfg.Notify.Overdue {
					s.mailTicket(ctx, tx, &nb, org, *r.handler, "overdue", &Ticket{ID: r.id, Ref: r.ref}, nil, "", "")
					if err := nb.Add(ctx, tx, org, *r.handler, "ticket", r.ref+" is overdue.", "#FF9B8A", r.ref); err != nil {
						return err
					}
				}
				late++
			} else if r.handler != nil && cfg.Notify.Nearing {
				if !told[r.handler.String()+"|"+r.ref] {
					s.mailTicket(ctx, tx, &nb, org, *r.handler, "near", &Ticket{ID: r.id, Ref: r.ref}, nil, "", "")
					if err := nb.Add(ctx, tx, org, *r.handler, "ticket-near", r.ref+" is nearing its deadline.", "#FFB86B", r.ref); err != nil {
						return err
					}
					near++
				}
			}
		}
		return nil
	})
	if err == nil {
		nb.Send(ctx, s.A)
	}
	return
}

// SweepAll sweeps every active organization (the worker calls this on a timer).
func (s *Service) SweepAll(ctx context.Context) error {
	var orgs []uuid.UUID
	if err := s.A.DB.Tx(ctx, func(tx pgx.Tx) error {
		rs, err := tx.Query(ctx, `select id from hub.orgs where status='active'`)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var id uuid.UUID
			if err := rs.Scan(&id); err != nil {
				return err
			}
			orgs = append(orgs, id)
		}
		return rs.Err()
	}); err != nil {
		return err
	}
	sort.Slice(orgs, func(i, j int) bool { return orgs[i].String() < orgs[j].String() })
	for _, o := range orgs {
		if _, _, err := s.Sweep(ctx, o); err != nil {
			return err
		}
	}
	return nil
}

// LoadConfig returns the organization's ticket timing policy (other modules use its working hours).
func LoadConfig(ctx context.Context, tx pgx.Tx) (Config, error) { return (&Service{}).config(ctx, tx) }

// isHead says whether a person heads the place a ticket is with: they hold a role that heads one place ("Each department", "Each
// sector" ...) and are placed in that place or in one above it (the head of a sector heads its branches too).
func isHead(ctx context.Context, tx pgx.Tx, person uuid.UUID, t *Ticket) (bool, error) {
	if t.ToRef == nil || t.ToType == "" || t.ToType == "desk" {
		return false, nil
	}
	var head bool
	err := tx.QueryRow(ctx, `with recursive up(id, parent_id) as (select id, parent_id from hub.org_units where id=$2
			union all select u.id, u.parent_id from hub.org_units u join up on u.id = up.parent_id)
		select exists (select 1 from hub.people pe join hub.person_roles pr on pr.person_id = pe.id join hub.roles r on r.id = pr.role_id
			where pe.id = $1 and exists (select 1 from up where up.id in (pe.sector_id, pe.branch_id, pe.unit_id, pe.department_id)
			  -- a role that heads each place of its kind, or the head role set for this place (or one above it) in Organization structure
			  and coalesce(hub.chosen_head(up.id) = pe.id, r.scope like 'Each %' or exists (select 1 from hub.unit_heads uh where uh.unit_id = up.id and uh.role_id = r.id))))`, person, *t.ToRef).Scan(&head)
	return head, err
}

// mayAssign: a head of the ticket's place assigns and hands over. Only heads do, for a ticket with a department, unit, branch or
// sector; at the ticket desk (no head) someone whose role holds "Assign & hand over" does, when they did not raise it.
func mayAssign(ctx context.Context, tx pgx.Tx, person uuid.UUID, t *Ticket) (bool, error) {
	head, err := isHead(ctx, tx, person, t)
	if err != nil || head {
		return head, err
	}
	if t.ToRef != nil && t.ToType != "desk" {
		// the organization's owner (a locked role) oversees every place, except tickets they raised; anyone else must head it
		var owner bool
		err := tx.QueryRow(ctx, `select exists (select 1 from hub.person_roles pr join hub.roles r on r.id = pr.role_id where pr.person_id=$1 and r.locked)`, person).Scan(&owner)
		return owner && t.RaisedBy != person, err // never on a ticket they raised themselves
	}
	if t.RaisedBy == person {
		return false, nil
	}
	return authz.Can(ctx, tx, person, "tickets", "Assign & hand over")
}

// AutoCloseAfter is how long a resolved ticket waits for the person who raised it to close it before Hub closes it.
const AutoCloseAfter = 30 * time.Minute

// autoClose closes tickets resolved more than AutoCloseAfter ago that nobody closed: rated 5 stars, with a comment from Hub saying so.
func (s *Service) autoClose(ctx context.Context, tx pgx.Tx, org uuid.UUID, nb *notify.Batch) error {
	rs, err := tx.Query(ctx, `select conversation_id, ref, raised_by from hub.tickets where step = 3 and solved_at is not null and solved_at < now() - make_interval(secs => $1) for update skip locked`, AutoCloseAfter.Seconds())
	if err != nil {
		return err
	}
	type due struct {
		id, raiser uuid.UUID
		ref        string
	}
	var ids []due
	for rs.Next() {
		var d due
		if err := rs.Scan(&d.id, &d.ref, &d.raiser); err != nil {
			rs.Close()
			return err
		}
		ids = append(ids, d)
	}
	rs.Close()
	if err := rs.Err(); err != nil {
		return err
	}
	mins := int(AutoCloseAfter.Minutes())
	for _, d := range ids {
		id := d.id
		tag, err := tx.Exec(ctx, `update hub.tickets set step=4, closed_at=now(), rating=5, auto_closed=true, updated_at=now(), `+integrity.Bump+` where conversation_id=$1 and step=3`, id)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			continue
		}
		if err := s.A.Sign.Seal(ctx, tx, "tickets", "conversation_id=$1", id); err != nil {
			return err
		}
		if err := s.addComment(ctx, tx, org, id, nil, fmt.Sprintf("Closed automatically %d minutes after it was resolved, as nobody closed it. Rated 5 stars.", mins)); err != nil {
			return err
		}
		if err := s.event(ctx, tx, org, id, nil, "event", fmt.Sprintf("Closed automatically %d minutes after it was resolved (rated 5 of 5)", mins)); err != nil {
			return err
		}
		if err := nb.Add(ctx, tx, org, d.raiser, "ticket-closed", d.ref+" was closed automatically and rated 5 stars.", "#B8F36B", d.ref); err != nil {
			return err
		}
	}
	return nil
}

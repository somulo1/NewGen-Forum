package tickets

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/authz"
	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/notify"
	"hub/internal/tenant"
)

// Act changes a ticket. Every change leaves an event in the ticket's history and a record in the audit log.
func (s *Service) Act(ctx context.Context, p *tenant.Principal, id uuid.UUID, a Action) (*Ticket, error) {
	var out *Ticket
	var nb notify.Batch
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		cur, err := scanTicket(tx.QueryRow(ctx, ticketSelect+` where t.conversation_id=$1 for update of t`, id))
		if errors.Is(err, pgx.ErrNoRows) {
			return httpx.ErrNotFound
		}
		if err != nil {
			return err
		}
		cfg, err := s.config(ctx, tx)
		if err != nil {
			return err
		}
		isHandler := cur.Handler != nil && *cur.Handler == p.Person
		raisedByMe := cur.RaisedBy == p.Person
		// the head of the place it is with now (its dashboard received it; or a place above that one) works on it as of right:
		// resolves, closes and moves it along, whatever their role's ticket rights. Not what their place sent or passed on.
		headHere, err := isHead(ctx, tx, p.Person, cur)
		if err != nil {
			return err
		}
		need := func(right string) error {
			// a head deals with what reached their place, even what they sent to it themselves
			if headHere && (right == "Solve" || right == "Close" || right == "Assign & hand over") {
				return nil
			}
			// anyone else never works on a ticket they raised; they may still close it
			if right == "Solve" && raisedByMe {
				return httpx.Err(403, "own_ticket", "You raised this ticket, so someone else has to work on it. You can still close it.")
			}
			if isHandler && (right == "Solve" || right == "Close") {
				return nil
			}
			if headHere && (right == "Solve" || right == "Close" || right == "Assign & hand over") {
				return nil
			}
			if right == "Close" { // the person who raised it may close it with "Close own tickets"
				return authz.RequireOwn(ctx, tx, p.Person, "tickets", "Close", "Close own tickets", cur.RaisedBy == p.Person)
			}
			return authz.Require(ctx, tx, p.Person, "tickets", right)
		}
		// Who may assign, reassign and escalate: the head of the place the ticket is with (or of a place above it), or someone whose role
		// holds "Assign & hand over" (Roles & rights). The person who raised it may not, unless they are such a head. The person
		// working on it may also escalate, but not hand it to someone else.
		mayAssign, err := mayAssign(ctx, tx, p.Person, cur)
		if err != nil {
			return err
		}
		needHead := func(what string) error {
			if mayAssign {
				return nil
			}
			return httpx.Err(403, "not_head", "Only the head of the department this ticket is with, or someone whose role may assign tickets, can "+what+" it.")
		}
		// a closed ticket is final: nothing changes it any more (it can only be reopened while resolved and not yet closed)
		// a closed ticket is final where it is; it can still be forwarded to another place, which opens it there on its own clock
		if cur.Step >= 4 && !(a.Do == "escalate" && a.ToType != "") {
			return httpx.Err(409, "ticket_closed", "This ticket is closed. It cannot be reopened or changed here; forward it to another place instead.")
		}
		set, text := "", ""
		// what happens to the stays once the ticket row is updated (see hops.go)
		var hopOp string
		var hopTo struct {
			typ string
			ref *uuid.UUID
			cat string
			tat int
		}
		args := []any{id}
		switch a.Do {
		case "assign":
			if err := needHead("assign or reassign"); err != nil {
				return err
			}
			if a.Handler == nil {
				return httpx.Bad("Choose who takes it.")
			}
			if *a.Handler == cur.RaisedBy {
				return httpx.Bad("The person who raised a ticket cannot be the one to work on it. Choose someone else.")
			}
			var name string
			if err := tx.QueryRow(ctx, `select first_name || ' ' || last_name from hub.people p where id=$1 and status='active' and exists (select 1 from hub.conversation_members m where m.conversation_id=$2 and m.account_id=p.account_id and m.left_at is null)`, *a.Handler, id).Scan(&name); err != nil {
				return httpx.Bad("That person cannot take this ticket: they are not part of its conversation. Add them first.")
			}
			args = append(args, *a.Handler)
			set = "handler=$2, handler_since=now(), step=greatest(step,1)"
			text = "Assigned it to " + name
			if _, err := tx.Exec(ctx, `insert into hub.ticket_assignees (org_id, ticket_id, person_id) values ($1,$2,$3) on conflict do nothing`, p.Org, id, *a.Handler); err != nil {
				return err
			}
			if err := s.A.Sign.Seal(ctx, tx, "ticket_assignees", "ticket_id=$1", id); err != nil {
				return err
			}
			if cfg.Notify.Assigned && *a.Handler != p.Person {
				s.mailTicket(ctx, tx, &nb, p.Org, *a.Handler, "handover", cur, a.Details.clean(), p.Name, "")
				if err := nb.Add(ctx, tx, p.Org, *a.Handler, "ticket", cur.Ref+" was handed over to you.", "#7CC4FF", cur.Ref); err != nil {
					return err
				}
			}
		case "step":
			if a.Step == nil || *a.Step < 0 || *a.Step > 4 {
				return httpx.Bad("Unknown step.")
			}
			right := map[int]string{2: "Solve", 3: "Solve", 4: "Close", 1: "Assign & hand over", 0: "Assign & hand over"}[*a.Step]
			if err := need(right); err != nil {
				return err
			}
			if *a.Step == 4 && cur.Priority == "High" {
				// closing a High priority ticket follows flow f5; handled by approval in the caller when it needs levels
				var lv int
				_ = tx.QueryRow(ctx, `select levels from hub.approval_flows where key='f5'`).Scan(&lv)
				if lv > 0 && cur.Step != 3 {
					return httpx.Err(409, "needs_approval", "Closing a High priority ticket needs approval. Resolve it first.")
				}
			}
			args = append(args, *a.Step)
			set = "step=$2::smallint, hold=false, solved_at = case when $2::int >= 3 then coalesce(solved_at, now()) else null end"
			switch {
			case cur.Step < 3 && *a.Step >= 3:
				hopOp = "resolve"
			case cur.Step >= 3 && *a.Step < 3:
				hopOp = "reopen"
			}
			text = "Moved it to " + stepNames[*a.Step]
			note := strings.TrimSpace(a.Comment)
			if len(note) > 2000 {
				return httpx.Bad("Keep the comment under 2000 characters.")
			}
			switch *a.Step {
			case 3: // resolving needs a word on what was done
				if note == "" {
					return httpx.Bad("Write what was done to resolve it.")
				}
				if err := s.addComment(ctx, tx, p.Org, id, &p.Person, "Resolved: "+note); err != nil {
					return err
				}
			case 4: // closing: optional stars, and a comment like every action
				if a.Rating < 0 || a.Rating > 5 {
					return httpx.Bad("Rate it from 1 to 5 stars.")
				}
				if a.Rating == 0 {
					return httpx.Bad("Rate it from 1 to 5 stars: every closure is rated.")
				}
				if note == "" {
					return httpx.Bad("Write a comment on closing it.")
				}
				args = append(args, a.Rating)
				set += fmt.Sprintf(", closed_at=now(), rating=nullif($%d::int, 0)::smallint, auto_closed=false", len(args))
				if a.Rating > 0 {
					text += fmt.Sprintf(" (rated %d of 5)", a.Rating)
				}
				if note != "" {
					if err := s.addComment(ctx, tx, p.Org, id, &p.Person, "Closed: "+note); err != nil {
						return err
					}
				}
			default: // reopened or moved back: it is no longer closed
				set += ", closed_at=null, rating=null, auto_closed=false"
			}
			if *a.Step == 3 && cfg.Notify.Solved {
				s.mailTicket(ctx, tx, &nb, p.Org, cur.RaisedBy, "resolved", cur, a.Details.clean(), p.Name, "")
				if err := nb.Add(ctx, tx, p.Org, cur.RaisedBy, "ticket", cur.Ref+" was resolved.", "#B8F36B", cur.Ref); err != nil {
					return err
				}
			}
		case "hold":
			if err := need("Solve"); err != nil {
				return err
			}
			if cur.Step != 2 {
				return httpx.Bad("Only a ticket in progress can be put on hold.")
			}
			set, text = "hold=true, hold_since=now()", "Put it on hold"
		case "resume":
			if err := need("Solve"); err != nil {
				return err
			}
			if !cur.Hold {
				return httpx.Bad("It is not on hold.")
			}
			// the time spent on hold does not count against the TAT
			if cfg.Pause.Onhold && cur.DueAt != nil {
				var since *time.Time
				_ = tx.QueryRow(ctx, `select hold_since from hub.tickets where conversation_id=$1`, id).Scan(&since)
				if since != nil {
					args = append(args, cfg.Due(*cur.DueAt, int(cfg.Running(*since, app.Now()).Minutes()/60)))
					set = "hold=false, hold_since=null, due_at=$2"
				}
			}
			if set == "" {
				set = "hold=false, hold_since=null"
			}
			text = "Resumed work"
		case "escalate":
			if !isHandler && !mayAssign {
				return httpx.Err(403, "not_head", "Only the person working on this ticket or the head of the department it is with can escalate it.")
			}
			if cur.Step >= 3 && a.ToType == "" {
				return httpx.Bad("It is already resolved. Forward it to another place if it needs more work.")
			}
			if cur.Escalated && a.ToType == "" {
				return httpx.Err(409, "already_escalated", "This ticket was already escalated.")
			}
			a.Why = strings.TrimSpace(a.Why)
			if len(a.Why) > 300 {
				return httpx.Bad("Keep the reason under 300 characters.")
			}
			if a.Why == "" {
				return httpx.Bad("Say why you are escalating it.")
			}
			set, text = "escalated=true", "Escalated to "+cfg.Escalate
			if a.ToType == "" && a.Person == nil && a.Category == "" {
				// reporting it (a missed or threatened TAT) to the heads of the place it is with; it stays there
				heads, err := s.reportMissedTAT(ctx, tx, &nb, p.Org, cur)
				if err != nil {
					return err
				}
				text = "Reported it to the head of the place it is with"
				if heads == 0 {
					text = "Reported it to " + cfg.Escalate
					if err := s.notifyEscalation(ctx, tx, &nb, p.Org, cur, cfg); err != nil {
						return err
					}
				}
				if a.Why != "" {
					text += ": " + strings.TrimSpace(a.Why)
				}
				break
			}
			dest := "the ticket desk"
			if a.ToType != "" {
				// it moves to another place: that place's clock starts now, on the category it falls under there
				switch cur.Kind {
				case "request":
					if a.ToType != "department" {
						return httpx.Bad("A business request can only go to another department.")
					}
					a.Category = RequestCategory
				case "communication":
					if a.ToType == "desk" {
						return httpx.Bad("A communication goes to a department, unit, branch or sector.")
					}
					a.Category = CommunicationCategory
				}
				if a.Category == "" {
					// no category chosen: the one it has, if the place it goes to handles it (checked below); its clock still restarts there
					a.Category = cur.Category
				}
				if a.ToType != "desk" && a.ToRef != nil && cur.ToRef != nil && *a.ToRef == *cur.ToRef && a.ToType == cur.ToType {
					return httpx.Bad("It is already there. Choose another place to escalate it to.")
				}
				if a.ToType == "desk" {
					args = append(args, "desk", (*uuid.UUID)(nil))
				} else {
					if a.ToRef == nil {
						return httpx.Bad("Choose where to escalate it.")
					}
					var name string
					if err := tx.QueryRow(ctx, `select name from hub.org_units where id=$1 and kind=$2`, *a.ToRef, a.ToType).Scan(&name); err != nil {
						return httpx.Bad("That " + a.ToType + " does not exist.")
					}
					if err := escalationAllowed(ctx, tx, p.Person, cur, a.ToType, *a.ToRef, name); err != nil {
						return err
					}
					if err := checkIntakeEscalation(ctx, tx, p.Person, *a.ToRef, name); err != nil {
						return err
					}
					dest = "the " + a.ToType + " " + name
					args = append(args, a.ToType, a.ToRef)
				}
				// remember the unit it leaves (the right side reads the row before this change)
				set += fmt.Sprintf(", escalated_from=case when to_ref is null then escalated_from else array_append(coalesce(escalated_from, '{}'), to_ref) end, to_type=$%d, to_ref=$%d", len(args)-1, len(args))
			}
			if a.Category != "" && a.ToType == "" {
				// a new category restarts the clock, so it only comes with a move to another place; staying put keeps its turnaround
				return httpx.Bad("Choose where to escalate it. The category is the one it falls under there.")
			}
			if a.Category != "" {
				tt, tr := cur.ToType, cur.ToRef
				if a.ToType != "" {
					tt, tr = a.ToType, a.ToRef
				}
				cat, err := s.category(ctx, tx, tt, tr, a.Category)
				if err != nil {
					return err
				}
				if cat == nil && (a.Category == RequestCategory || a.Category == CommunicationCategory) {
					cat = &Category{UnitType: tt, UnitRef: tr, Name: a.Category, TAT: DefaultKindTAT(map[string]string{RequestCategory: "request", CommunicationCategory: "communication"}[a.Category])}
				}
				if cat == nil {
					return httpx.Bad("That category does not exist for the place you are escalating to.")
				}
				at := app.Now()
				if a.At != nil {
					at = *a.At
				}
				due := cfg.Due(at, cat.TAT)
				args = append(args, cat.Name, cat.TAT, due, PriorityOf(cat.TAT))
				set += fmt.Sprintf(", category=$%d, tat_hours=$%d, due_at=$%d, priority=$%d, late=false, hold=false, hold_since=null", len(args)-3, len(args)-2, len(args)-1, len(args))
				dest += " as " + cat.Name
				if a.ToType != "" {
					hopOp = "move"
					hopTo.typ, hopTo.ref, hopTo.cat, hopTo.tat = tt, tr, cat.Name, cat.TAT
					// it arrives open at the new place, even if it was resolved or closed where it was
					set += ", solved_at=null, closed_at=null, rating=null, auto_closed=false"
					if a.Person == nil {
						set += ", handler=null, handler_since=null, step=0"
					} else {
						set += ", step=1"
					}
				}
			}
			if a.Person != nil {
				if *a.Person == cur.RaisedBy {
					return httpx.Bad("The person who raised a ticket cannot be the one to work on it. Choose someone else.")
				}
				var name string
				if err := tx.QueryRow(ctx, `select first_name || ' ' || last_name from hub.people p where id=$1 and status='active' and exists (select 1 from hub.conversation_members m where m.conversation_id=$2 and m.account_id=p.account_id and m.left_at is null)`, *a.Person, id).Scan(&name); err != nil {
					return httpx.Bad("That person cannot take this ticket: they are not part of its conversation. Add them first.")
				}
				args = append(args, *a.Person)
				if hopOp == "move" {
					set += fmt.Sprintf(", handler=$%d, handler_since=now()", len(args)) // its step is set with the move
				} else {
					set += fmt.Sprintf(", handler=$%d, handler_since=now(), step=greatest(step,1)", len(args))
				}
				dest += " (" + name + ")"
				if _, err := tx.Exec(ctx, `insert into hub.ticket_assignees (org_id, ticket_id, person_id) values ($1,$2,$3) on conflict do nothing`, p.Org, id, *a.Person); err != nil {
					return err
				}
				if err := s.A.Sign.Seal(ctx, tx, "ticket_assignees", "ticket_id=$1", id); err != nil {
					return err
				}
				s.mailTicket(ctx, tx, &nb, p.Org, *a.Person, "escalated", cur, a.Details.clean(), p.Name, a.Why)
				if err := nb.Add(ctx, tx, p.Org, *a.Person, "ticket", cur.Ref+" was escalated to you.", "#FF9B8A", cur.Ref); err != nil {
					return err
				}
			}
			text = "Escalated to " + dest
			if a.Why != "" {
				text += ": " + a.Why
			}
		case "priority":
			if err := needHead("change the priority of"); err != nil {
				return err
			}
			if a.Priority != "Low" && a.Priority != "Medium" && a.Priority != "High" {
				return httpx.Bad("Choose Low, Medium or High.")
			}
			args = append(args, a.Priority)
			set, text = "priority=$2", "Changed the priority to "+a.Priority
		default:
			return httpx.Bad("Unknown action.")
		}
		if _, err := tx.Exec(ctx, `update hub.tickets set `+set+`, updated_at=now(), `+integrity.Bump+` where conversation_id=$1`, args...); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "tickets", "conversation_id=$1", id); err != nil {
			return err
		}
		hopAt := app.Now()
		if a.At != nil {
			hopAt = *a.At
		}
		switch hopOp {
		case "move":
			if err := s.closeHop(ctx, tx, p.Org, cur, "escalated", hopAt); err != nil {
				return err
			}
			if err := s.openHop(ctx, tx, p.Org, id, hopTo.typ, hopTo.ref, hopTo.cat, hopTo.tat, hopAt); err != nil {
				return err
			}
			if a.Person == nil {
				// nobody named: whoever the new place's category (or its heads) names takes it, if they are on it
				if now, err := scanTicket(tx.QueryRow(ctx, ticketSelect+` where t.conversation_id=$1`, id)); err == nil {
					if _, err := s.autoAssign(ctx, tx, &nb, p.Org, now); err != nil {
						return err
					}
				}
			}
		case "resolve":
			if err := s.closeHop(ctx, tx, p.Org, cur, "resolved", hopAt); err != nil {
				return err
			}
		case "reopen":
			if err := s.reopenHop(ctx, tx, cur); err != nil {
				return err
			}
		}
		if err := s.eventAt(ctx, tx, p.Org, id, p, "event", text, a.At); err != nil {
			return err
		}
		if err := s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, cur.Ref+": "+text, "Tickets", p.IP, map[string]any{"ticket": id, "do": a.Do}); err != nil {
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

// escalationAllowed checks where a ticket may be escalated to. It must move somewhere new, and a person may not send it to
// their own sector, branch, unit or department at the level it already sits at (their unit's ticket into their own unit,
// their branch's into their own branch). Moving it a level up into their own place is fine: from their unit into their
// branch, from their branch into their sector.
func escalationAllowed(ctx context.Context, tx pgx.Tx, person uuid.UUID, cur *Ticket, toType string, toRef uuid.UUID, name string) error {
	if cur.ToType == toType && cur.ToRef != nil && *cur.ToRef == toRef {
		return httpx.Bad("It is already with the " + toType + " " + name + ". Choose another place.")
	}
	var own bool
	if err := tx.QueryRow(ctx, `select coalesce($2 in (sector_id, branch_id, unit_id, department_id), false) from hub.people where id=$1`, person, toRef).Scan(&own); err != nil {
		return err
	}
	if own && cur.ToType == toType {
		return httpx.Bad("You cannot escalate a " + toType + " ticket into your own " + toType + ". Escalate it to a higher level, or to another " + toType + ".")
	}
	return nil
}

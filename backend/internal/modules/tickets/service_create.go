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
	"hub/internal/notify"
	"hub/internal/tenant"
)

// Create turns a ticket conversation (made by the requester's device, with a key for the requester and the people who may
// handle it) into a ticket: it picks a handler, starts the clock and tells the handler.
func (s *Service) Create(ctx context.Context, p *tenant.Principal, in CreateIn) (*Ticket, error) {
	if in.Priority != "" && in.Priority != "Low" && in.Priority != "Medium" && in.Priority != "High" {
		return nil, httpx.Bad("Choose Low, Medium or High.")
	}
	if in.ToType == "" {
		in.ToType = "desk"
	}
	if in.ToType != "desk" && in.ToType != "department" && in.ToType != "unit" && in.ToType != "branch" && in.ToType != "sector" {
		return nil, httpx.Bad("Unknown recipient.")
	}
	if in.ToType != "desk" && in.ToRef == nil {
		return nil, httpx.Bad("Choose who receives it.")
	}
	if in.Kind != "" && in.Kind != "request" && in.Kind != "ticket" && in.Kind != "communication" {
		return nil, httpx.Bad("Choose a ticket, a business request or a communication.")
	}
	if in.Kind == "communication" && (in.ToType == "desk" || in.ToRef == nil) {
		return nil, httpx.Bad("A communication goes to a department, unit, branch or sector.")
	}
	if in.Kind == "communication" && in.FromRef == nil {
		return nil, httpx.Bad("A communication is sent from a department, unit, branch or sector dashboard.")
	}
	if in.FromRef != nil && in.Kind == "" {
		return nil, httpx.Bad("Tickets are raised by people, not on behalf of a place.")
	}
	if in.FromRef != nil && in.ToRef != nil && *in.FromRef == *in.ToRef {
		return nil, httpx.Bad("Choose another place to send it to.")
	}
	if in.Kind == "ticket" {
		in.Kind = ""
	}
	if in.Kind == "request" && in.ToType != "department" {
		return nil, httpx.Bad("Business requests go to a department.")
	}
	if in.Kind == "request" && in.FromRef != nil && in.FromType != "" && in.FromType != "department" {
		return nil, httpx.Bad("Business requests are sent from a department.")
	}
	var err error
	if in.Title, err = cleanText(in.Title, maxTitle, "The subject"); err != nil {
		return nil, err
	}
	// a ticket's or business request's subject is 250 words at most
	if len(strings.Fields(in.Title)) > maxTitleWords {
		return nil, httpx.Bad(fmt.Sprintf("Keep the subject to %d words or fewer.", maxTitleWords))
	}
	if in.Description, err = cleanText(in.Description, maxDesc, "The description"); err != nil {
		return nil, err
	}
	var out *Ticket
	var nb notify.Batch
	err = s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "tickets", "Raise"); err != nil {
			return err
		}
		var kind string
		var creator uuid.UUID
		if err := tx.QueryRow(ctx, `select c.kind, c.created_by from hub.conversations c join hub.conversation_members m on m.conversation_id=c.id and m.account_id=$2 and m.left_at is null where c.id=$1`, in.Conversation, p.Account).Scan(&kind, &creator); err != nil {
			return httpx.ErrNotFound
		}
		if kind != "ticket" || creator != p.Account {
			return httpx.Bad("Create the ticket conversation first.")
		}
		// business requests are a departmental matter: only people with access to departments send them
		if in.Kind == "request" {
			if err := authz.Require(ctx, tx, p.Person, "departments", "View dashboard"); err != nil {
				return httpx.Err(403, "no_department_access", "Only people with access to departments can send business requests.")
			}
		}
		// a place that only takes escalations from one chosen place cannot be sent tickets or business requests directly
		if in.ToRef != nil {
			if closed, err := intakeClosed(ctx, tx, *in.ToRef); err != nil {
				return err
			} else if closed {
				return httpx.Err(403, "intake_closed", "That place only takes tickets escalated from places it chose. Send it somewhere else.")
			}
		}
		if in.FromRef != nil {
			// only someone who belongs to the place (or heads it) and may open its dashboard sends on its behalf
			var fromKind string
			var placed bool
			if err := tx.QueryRow(ctx, `select u.kind, exists (select 1 from hub.people pe where pe.id=$2 and $1 in (pe.department_id, pe.unit_id, pe.branch_id, pe.sector_id))
				from hub.org_units u where u.id=$1`, *in.FromRef, p.Person).Scan(&fromKind, &placed); err != nil {
				return httpx.Bad("That place does not exist.")
			}
			if !placed {
				return httpx.Err(403, "not_member", "You can only send on behalf of a place you belong to.")
			}
			mod := map[string]string{"department": "departments", "unit": "units", "branch": "branches", "sector": "sectors"}[fromKind]
			if err := authz.Require(ctx, tx, p.Person, mod, "View dashboard"); err != nil {
				return err
			}
			in.FromType = fromKind
			if in.Kind == "request" && fromKind != "department" {
				return httpx.Bad("Business requests are sent from a department.")
			}
		}
		if in.Kind == "request" || in.Kind == "communication" {
			in.Category = KindCategory(in.Kind) // its turnaround is the one the receiving place set for these
		}
		cat, err := s.category(ctx, tx, in.ToType, in.ToRef, in.Category)
		if err != nil {
			return err
		}
		if cat == nil && (in.Kind == "request" || in.Kind == "communication") {
			cat = &Category{UnitType: in.ToType, UnitRef: in.ToRef, Name: KindCategory(in.Kind), TAT: DefaultKindTAT(in.Kind), Members: []uuid.UUID{}}
		}
		if cat == nil {
			return httpx.Bad("That category does not exist for this recipient.")
		}
		// priority follows how fast the department promised to resolve it: what is due within a working day is urgent
		in.Priority = PriorityOf(cat.TAT)
		cfg, err := s.config(ctx, tx)
		if err != nil {
			return err
		}
		now := app.Now()
		if in.At != nil {
			now = in.At.UTC().Truncate(time.Microsecond)
		}
		due := cfg.Due(now, cat.TAT)
		handler, err := s.pickHandler(ctx, tx, cat)
		if err != nil {
			return err
		}
		var prefix string
		if in.ToRef != nil {
			_ = tx.QueryRow(ctx, `select coalesce(nullif(upper(left(code,4)),''), upper(left(name,2))) from hub.org_units where id=$1`, *in.ToRef).Scan(&prefix)
		}
		if prefix == "" {
			prefix = "TK"
		}
		var n int
		if err := tx.QueryRow(ctx, `insert into hub.counters (org_id, name, value) values ($1,'ticket:'||$2,100) on conflict (org_id, name) do update set value = hub.counters.value + 1 returning value`, p.Org, prefix).Scan(&n); err != nil {
			return err
		}
		ref := fmt.Sprintf("%s-%d", prefix, n)
		if in.Ref != "" {
			ref = in.Ref
		}
		step := 0
		var since *time.Time
		if handler != nil {
			step, since = 1, &now
		}
		if _, err := tx.Exec(ctx, `insert into hub.tickets (org_id, conversation_id, ref, category, priority, step, raised_by, handler, handler_since, to_type, to_ref, tat_hours, due_at, sync, created_at, kind, from_type, from_ref)
			values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,nullif($16,''),nullif($17,''),$18)`, p.Org, in.Conversation, ref, cat.Name, in.Priority, step, p.Person, handler, since, in.ToType, in.ToRef, cat.TAT, due, in.Sync, now, in.Kind, in.FromType, in.FromRef); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "tickets", "conversation_id=$1", in.Conversation); err != nil {
			return err
		}
		if handler == nil {
			// nobody named by the category is on it yet: the heads of the place (a business request or communication has no list)
			t0 := &Ticket{ID: in.Conversation, Ref: ref, Kind: map[string]string{"": "ticket"}[in.Kind] + in.Kind, Category: cat.Name, ToType: in.ToType, ToRef: in.ToRef, RaisedBy: p.Person, Step: step}
			if _, err := s.autoAssign(ctx, tx, &nb, p.Org, t0); err != nil {
				return err
			}
		}
		// its first stay: the receiving place's clock starts now
		if err := s.openHop(ctx, tx, p.Org, in.Conversation, in.ToType, in.ToRef, cat.Name, cat.TAT, now); err != nil {
			return err
		}
		if in.Title != "" {
			if err := s.saveText(ctx, tx, p.Org, in.Conversation, in.Title, in.Description); err != nil {
				return err
			}
		}
		raisedText := map[string]string{"": "Raised the ticket", "request": "Raised the business request", "communication": "Sent the communication"}[in.Kind]
		if err := s.eventAt(ctx, tx, p.Org, in.Conversation, p, "event", raisedText, &now); err != nil {
			return err
		}
		if handler != nil {
			if _, err := tx.Exec(ctx, `insert into hub.ticket_assignees (org_id, ticket_id, person_id) values ($1,$2,$3)`, p.Org, in.Conversation, *handler); err != nil {
				return err
			}
			if err := s.A.Sign.Seal(ctx, tx, "ticket_assignees", "ticket_id=$1", in.Conversation); err != nil {
				return err
			}
			if err := s.eventAt(ctx, tx, p.Org, in.Conversation, nil, "event", "Assigned automatically to the handler with the fewest open tickets", &now); err != nil {
				return err
			}
			if cfg.Notify.Assigned {
				s.mailTicket(ctx, tx, &nb, p.Org, *handler, "assigned", &Ticket{ID: in.Conversation, Ref: ref, Category: cat.Name, Priority: in.Priority, RaisedByName: p.Name, DueAt: &due, StepName: stepNames[step]}, in.Details.clean(), p.Name, "")
				if err := nb.Add(ctx, tx, p.Org, *handler, "ticket", ref+" was assigned to you.", "#FF9B8A", ref); err != nil {
					return err
				}
			}
		}
		out, err = s.get(ctx, tx, p.Org, in.Conversation)
		return err
	})
	if err != nil {
		return nil, err
	}
	nb.Send(ctx, s.A)
	s.changed(ctx, p, out)
	return out, nil
}

// pickHandler: the member of the category with the fewest open tickets takes the new one.
func (s *Service) pickHandler(ctx context.Context, tx pgx.Tx, c *Category) (*uuid.UUID, error) {
	if len(c.Members) == 0 {
		return nil, nil
	}
	rs, err := tx.Query(ctx, `select p.id, (select count(*) from hub.tickets t where t.handler = p.id and t.step < 3) as load
		from hub.people p where p.id = any($1) and p.status='active' order by load, p.id`, c.Members)
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	if rs.Next() {
		var id uuid.UUID
		var load int
		if err := rs.Scan(&id, &load); err != nil {
			return nil, err
		}
		return &id, nil
	}
	return nil, rs.Err()
}

const ticketSelect = `select t.conversation_id, t.ref, coalesce(t.kind, 'ticket'), t.category, t.priority, t.step, t.hold, t.raised_by, coalesce(rp.first_name || ' ' || rp.last_name,''), t.handler, coalesce(hp.first_name || ' ' || hp.last_name,''),
	t.handler_since, t.to_type, t.to_ref, t.tat_hours, t.due_at, t.solved_at, t.escalated, t.late, t.sync, t.created_at, t.hold_since,
	coalesce((select array_agg(person_id) from hub.ticket_assignees a where a.ticket_id = t.conversation_id), '{}'), t.closed_at, t.rating, t.auto_closed
	from hub.tickets t join hub.people rp on rp.id = t.raised_by left join hub.people hp on hp.id = t.handler`

func scanTicket(r pgx.Row) (*Ticket, error) {
	t := &Ticket{}
	err := r.Scan(&t.ID, &t.Ref, &t.Kind, &t.Category, &t.Priority, &t.Step, &t.Hold, &t.RaisedBy, &t.RaisedByName, &t.Handler, &t.HandlerName, &t.HandlerSince, &t.ToType, &t.ToRef, &t.TATHours, &t.DueAt, &t.SolvedAt, &t.Escalated, &t.Late, &t.Sync, &t.CreatedAt, &t.holdSince, &t.Assignees, &t.ClosedAt, &t.Rating, &t.AutoClosed)
	if t.Assignees == nil {
		t.Assignees = []uuid.UUID{}
	}
	t.StepName = stepNames[min(max(t.Step, 0), 4)]
	if t.Step == 2 && t.Hold {
		t.StepName = "On hold"
	}
	return t, err
}

// withElapsed fills ElapsedHours from the due time, so it follows the same pauses the due time was built with.
func (c Config) withElapsed(t *Ticket) {
	if t.DueAt == nil {
		return
	}
	end := app.Now()
	if t.SolvedAt != nil {
		end = *t.SolvedAt
	} else if t.Hold && t.holdSince != nil && c.Pause.Onhold {
		end = *t.holdSince
	}
	tat := float64(t.TATHours)
	if end.Before(*t.DueAt) {
		t.ElapsedHours = max(0, tat-c.Running(end, *t.DueAt).Hours())
	} else {
		t.ElapsedHours = tat + c.Running(*t.DueAt, end).Hours()
	}
}

func (s *Service) get(ctx context.Context, tx pgx.Tx, org, id uuid.UUID) (*Ticket, error) {
	t, err := scanTicket(tx.QueryRow(ctx, ticketSelect+` where t.conversation_id=$1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, httpx.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if cfg, err := s.config(ctx, tx); err == nil {
		cfg.withElapsed(t)
	}
	as, err := tx.Query(ctx, `select pe.first_name || ' ' || pe.last_name, a.at from hub.ticket_acks a join hub.people pe on pe.id = a.person_id where a.ticket_id=$1 order by a.at`, id)
	if err != nil {
		return nil, err
	}
	for as.Next() {
		var a Ack
		if err := as.Scan(&a.Name, &a.At); err != nil {
			as.Close()
			return nil, err
		}
		t.Acks = append(t.Acks, a)
	}
	as.Close()
	rs, err := tx.Query(ctx, `select at, actor_name, kind, text from hub.ticket_events where ticket_id=$1 order by at, id`, id)
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	for rs.Next() {
		var e Event
		if err := rs.Scan(&e.At, &e.Actor, &e.Kind, &e.Text); err != nil {
			return nil, err
		}
		t.Events = append(t.Events, e)
	}
	return t, rs.Err()
}

type ListQuery struct {
	Mine, Open, Late bool
	Handler          *uuid.UUID
	Cursor           *time.Time
	Limit            int
}

// List shows ticket facts to anyone with the View right. Reading a ticket's text still needs a key, which only its
// participants hold.
func (s *Service) List(ctx context.Context, p *tenant.Principal, q ListQuery) ([]*Ticket, error) {
	if q.Limit <= 0 || q.Limit > 200 {
		q.Limit = 100
	}
	out := []*Ticket{}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "tickets", "View"); err != nil {
			return err
		}
		where, args := []string{"true"}, []any{}
		add := func(c string, v any) { args = append(args, v); where = append(where, fmt.Sprintf(c, len(args))) }
		if q.Mine {
			add("(t.raised_by = $%d or t.handler = $"+fmt.Sprint(len(args)+1)+" or exists (select 1 from hub.ticket_assignees a where a.ticket_id = t.conversation_id and a.person_id = $"+fmt.Sprint(len(args)+1)+"))", p.Person)
		}
		if q.Handler != nil {
			add("t.handler = $%d", *q.Handler)
		}
		if q.Open {
			where = append(where, "t.step < 3")
		}
		if q.Late {
			where = append(where, "t.late")
		}
		if q.Cursor != nil {
			add("t.created_at < $%d", *q.Cursor)
		}
		cfg, cerr := s.config(ctx, tx)
		rs, err := tx.Query(ctx, ticketSelect+" where "+strings.Join(where, " and ")+fmt.Sprintf(" order by t.created_at desc, t.conversation_id desc limit %d", q.Limit), args...)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			t, err := scanTicket(rs)
			if err != nil {
				return err
			}
			if cerr == nil {
				cfg.withElapsed(t)
			}
			out = append(out, t)
		}
		if err := rs.Err(); err != nil {
			return err
		}
		rs.Close()
		ids := make([]uuid.UUID, len(out))
		for i, t := range out {
			ids[i] = t.ID
		}
		in := map[uuid.UUID]bool{}
		ms, err := tx.Query(ctx, `select conversation_id from hub.conversation_members where account_id=$1 and left_at is null and conversation_id = any($2)`, p.Account, ids)
		if err != nil {
			return err
		}
		defer ms.Close()
		for ms.Next() {
			var id uuid.UUID
			if err := ms.Scan(&id); err != nil {
				return err
			}
			in[id] = true
		}
		var mine []*Ticket
		for _, t := range out {
			v := in[t.ID]
			t.Participant = &v
			if v {
				mine = append(mine, t)
			}
		}
		if err := ms.Err(); err != nil {
			return err
		}
		ms.Close()
		// attachment counts for the whole list in one query
		if len(mine) > 0 {
			ids := make([]uuid.UUID, len(mine))
			byID := map[uuid.UUID]*Ticket{}
			for i, t := range mine {
				ids[i] = t.ID
				byID[t.ID] = t
			}
			fc, err := tx.Query(ctx, `select ticket_id, count(*) from hub.ticket_files where ticket_id = any($1) group by ticket_id`, ids)
			if err != nil {
				return err
			}
			for fc.Next() {
				var id uuid.UUID
				var n int
				if err := fc.Scan(&id, &n); err != nil {
					fc.Close()
					return err
				}
				byID[id].FileCount = n
			}
			fc.Close()
		}
		return s.fillText(ctx, tx, p.Org, mine, false)
	})
	return out, err
}

// Ack is one acknowledgement of a business request or communication.
type Ack struct {
	Name string    `json:"name"`
	At   time.Time `json:"at"`
}

// mayAcknowledge: someone placed at the place a business request or communication is with (or above it) acknowledges receiving it.
func mayAcknowledge(ctx context.Context, tx pgx.Tx, person uuid.UUID, t *Ticket) (bool, error) {
	if t.Kind != "request" && t.Kind != "communication" || t.ToRef == nil {
		return false, nil
	}
	var ok bool
	err := tx.QueryRow(ctx, `with recursive up(id, parent_id) as (select id, parent_id from hub.org_units where id=$2
			union all select u.id, u.parent_id from hub.org_units u join up on u.id = up.parent_id)
		select exists (select 1 from hub.people pe where pe.id=$1 and exists (select 1 from up where up.id in (pe.department_id, pe.unit_id, pe.branch_id, pe.sector_id)))`, person, *t.ToRef).Scan(&ok)
	return ok, err
}

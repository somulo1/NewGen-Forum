// Package tickets is the ticket desk: routing, timing (TAT), hand-over and escalation. The subject, description and thread
// of a ticket are end-to-end encrypted in its conversation; this package handles everything the server must act on.
package tickets

import (
	"context"
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
	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/notify"
	"hub/internal/realtime"
	"hub/internal/tenant"
)

type Service struct{ A *app.App }

func New(a *app.App) *Service { return &Service{A: a} }

var stepNames = []string{"Raised", "Assigned", "In progress", "Resolved", "Closed"}

func (s *Service) config(ctx context.Context, tx pgx.Tx) (Config, error) {
	c := defaultConfig()
	var raw []byte
	err := tx.QueryRow(ctx, `select config from hub.ticket_settings`).Scan(&raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return c, nil
	}
	if err != nil {
		return c, err
	}
	_ = json.Unmarshal(raw, &c)
	return c, nil
}

type Category struct {
	ID       uuid.UUID   `json:"id"`
	UnitType string      `json:"unit_type"`
	UnitRef  *uuid.UUID  `json:"unit_ref,omitempty"`
	Name     string      `json:"name"`
	TAT      int         `json:"tat_hours"`
	Priority string      `json:"priority"`
	Members  []uuid.UUID `json:"members"`
}

type Ticket struct {
	ID           uuid.UUID   `json:"id"` // = its conversation
	Ref          string      `json:"ref"`
	Category     string      `json:"category"`
	Priority     string      `json:"priority"`
	Step         int         `json:"step"`
	StepName     string      `json:"step_name"`
	Hold         bool        `json:"hold"`
	RaisedBy     uuid.UUID   `json:"raised_by"`
	RaisedByName string      `json:"raised_by_name"`
	Handler      *uuid.UUID  `json:"handler,omitempty"`
	HandlerName  string      `json:"handler_name,omitempty"`
	HandlerSince *time.Time  `json:"handler_since,omitempty"`
	Assignees    []uuid.UUID `json:"assignees"`
	ToType       string      `json:"to_type"`
	ToRef        *uuid.UUID  `json:"to_ref,omitempty"`
	TATHours     int         `json:"tat_hours"`
	DueAt        *time.Time  `json:"due_at,omitempty"`
	SolvedAt     *time.Time  `json:"solved_at,omitempty"`
	Escalated    bool        `json:"escalated"`
	Late         bool        `json:"late"`
	Sync         string      `json:"sync,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
	Events       []Event     `json:"events,omitempty"`
	// ElapsedHours is how much of the TAT's running clock has been used (pauses and off-hours excluded); above TATHours means overdue.
	ElapsedHours float64 `json:"elapsed_hours"`
	holdSince    *time.Time
}

type Event struct {
	At    time.Time `json:"at"`
	Actor string    `json:"who"`
	Kind  string    `json:"kind"`
	Text  string    `json:"text"`
}

type CreateIn struct {
	Conversation uuid.UUID  `json:"conversation"`
	Category     string     `json:"category"`
	Priority     string     `json:"priority"`
	ToType       string     `json:"to_type"`
	ToRef        *uuid.UUID `json:"to_ref"`
	Sync         string     `json:"sync"`
	// At backdates the ticket (history loading only; not settable over HTTP).
	At *time.Time `json:"-"`
	// Ref fixes the ticket reference (history loading only).
	Ref string `json:"-"`
}

func (s *Service) event(ctx context.Context, tx pgx.Tx, org, ticket uuid.UUID, p *tenant.Principal, kind, text string) error {
	return s.eventAt(ctx, tx, org, ticket, p, kind, text, nil)
}

func (s *Service) eventAt(ctx context.Context, tx pgx.Tx, org, ticket uuid.UUID, p *tenant.Principal, kind, text string, at *time.Time) error {
	var actor *uuid.UUID
	name := "System"
	if p != nil {
		actor, name = &p.Person, p.Name
	}
	_, err := tx.Exec(ctx, `insert into hub.ticket_events (org_id, ticket_id, at, actor, actor_name, kind, text) values ($1,$2,coalesce($7, now()),$3,$4,$5,$6)`, org, ticket, actor, name, kind, text, at)
	return err
}

// categoriesFor finds the categories that apply to a receiving unit, falling back to the desk's.
func (s *Service) category(ctx context.Context, tx pgx.Tx, toType string, toRef *uuid.UUID, name string) (*Category, error) {
	c := &Category{}
	row := tx.QueryRow(ctx, `select id, unit_type, unit_ref, name, tat_hours, priority, members from hub.ticket_categories
		where name=$1 and ((unit_type=$2 and unit_ref is not distinct from $3) or unit_type='desk') order by (unit_type='desk') limit 1`, name, toType, toRef)
	if err := row.Scan(&c.ID, &c.UnitType, &c.UnitRef, &c.Name, &c.TAT, &c.Priority, &c.Members); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return c, nil
}

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
	var out *Ticket
	var nb notify.Batch
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
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
		cat, err := s.category(ctx, tx, in.ToType, in.ToRef, in.Category)
		if err != nil {
			return err
		}
		if cat == nil {
			return httpx.Bad("That category does not exist for this recipient.")
		}
		if in.Priority == "" {
			in.Priority = cat.Priority
		}
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
		if _, err := tx.Exec(ctx, `insert into hub.tickets (org_id, conversation_id, ref, category, priority, step, raised_by, handler, handler_since, to_type, to_ref, tat_hours, due_at, sync, created_at)
			values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`, p.Org, in.Conversation, ref, cat.Name, in.Priority, step, p.Person, handler, since, in.ToType, in.ToRef, cat.TAT, due, in.Sync, now); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "tickets", "conversation_id=$1", in.Conversation); err != nil {
			return err
		}
		if err := s.eventAt(ctx, tx, p.Org, in.Conversation, p, "event", "Raised the ticket", &now); err != nil {
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

const ticketSelect = `select t.conversation_id, t.ref, t.category, t.priority, t.step, t.hold, t.raised_by, coalesce(rp.first_name || ' ' || rp.last_name,''), t.handler, coalesce(hp.first_name || ' ' || hp.last_name,''),
	t.handler_since, t.to_type, t.to_ref, t.tat_hours, t.due_at, t.solved_at, t.escalated, t.late, t.sync, t.created_at, t.hold_since,
	coalesce((select array_agg(person_id) from hub.ticket_assignees a where a.ticket_id = t.conversation_id), '{}')
	from hub.tickets t join hub.people rp on rp.id = t.raised_by left join hub.people hp on hp.id = t.handler`

func scanTicket(r pgx.Row) (*Ticket, error) {
	t := &Ticket{}
	err := r.Scan(&t.ID, &t.Ref, &t.Category, &t.Priority, &t.Step, &t.Hold, &t.RaisedBy, &t.RaisedByName, &t.Handler, &t.HandlerName, &t.HandlerSince, &t.ToType, &t.ToRef, &t.TATHours, &t.DueAt, &t.SolvedAt, &t.Escalated, &t.Late, &t.Sync, &t.CreatedAt, &t.holdSince, &t.Assignees)
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
		return rs.Err()
	})
	return out, err
}

func (s *Service) Get(ctx context.Context, p *tenant.Principal, id uuid.UUID) (*Ticket, error) {
	var t *Ticket
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "tickets", "View"); err != nil {
			return err
		}
		var err error
		t, err = s.get(ctx, tx, p.Org, id)
		return err
	})
	return t, err
}

func (s *Service) changed(ctx context.Context, p *tenant.Principal, t *Ticket) {
	d, _ := json.Marshal(t)
	s.A.Publish(ctx, realtime.Event{Type: "ticket.updated", Channel: realtime.OrgChannel(p.Org), Org: p.Org.String(), Actor: p.Person.String(), Data: d})
	s.A.Publish(ctx, realtime.Event{Type: "ticket.updated", Channel: realtime.ConvChannel(t.ID), Org: p.Org.String(), Actor: p.Person.String(), Data: d})
}

type Action struct {
	Do        string      `json:"do"` // assign | step | hold | resume | escalate | priority
	Handler   *uuid.UUID  `json:"handler,omitempty"`
	Assignees []uuid.UUID `json:"assignees,omitempty"`
	Step      *int        `json:"step,omitempty"`
	Priority  string      `json:"priority,omitempty"`
	// escalate: where it goes (a desk, sector, branch, unit or department), who takes it there, and why
	ToType string     `json:"to_type,omitempty"`
	ToRef  *uuid.UUID `json:"to_ref,omitempty"`
	Person *uuid.UUID `json:"person,omitempty"`
	Why    string     `json:"why,omitempty"`
	// Category re-files the ticket under one of the destination's categories; its time limit starts again from now.
	Category string `json:"category,omitempty"`
	// At backdates the change (history loading only; not settable over HTTP).
	At *time.Time `json:"-"`
}

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
		need := func(right string) error {
			if isHandler && (right == "Solve" || right == "Close") {
				return nil
			}
			return authz.Require(ctx, tx, p.Person, "tickets", right)
		}
		set, text := "", ""
		args := []any{id}
		switch a.Do {
		case "assign":
			if err := need("Assign & hand over"); err != nil {
				return err
			}
			if a.Handler == nil {
				return httpx.Bad("Choose who takes it.")
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
			text = "Moved it to " + stepNames[*a.Step]
			if *a.Step == 3 && cfg.Notify.Solved {
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
			if err := need("Assign & hand over"); err != nil {
				return err
			}
			if cur.Escalated && a.ToType == "" {
				return httpx.Err(409, "already_escalated", "This ticket was already escalated.")
			}
			a.Why = strings.TrimSpace(a.Why)
			if len(a.Why) > 300 {
				return httpx.Bad("Keep the reason under 300 characters.")
			}
			set, text = "escalated=true", "Escalated to "+cfg.Escalate
			if a.ToType == "" && a.Person == nil && a.Category == "" {
				if err := s.notifyEscalation(ctx, tx, &nb, p.Org, cur, cfg); err != nil {
					return err
				}
				break
			}
			dest := "the ticket desk"
			if a.ToType != "" {
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
					dest = "the " + a.ToType + " " + name
					args = append(args, a.ToType, a.ToRef)
				}
				set += fmt.Sprintf(", to_type=$%d, to_ref=$%d", len(args)-1, len(args))
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
				if cat == nil {
					return httpx.Bad("That category does not exist for the place you are escalating to.")
				}
				due := cfg.Due(app.Now(), cat.TAT)
				args = append(args, cat.Name, cat.TAT, due)
				set += fmt.Sprintf(", category=$%d, tat_hours=$%d, due_at=$%d, late=false", len(args)-2, len(args)-1, len(args))
				dest += " as " + cat.Name
			}
			if a.Person != nil {
				var name string
				if err := tx.QueryRow(ctx, `select first_name || ' ' || last_name from hub.people p where id=$1 and status='active' and exists (select 1 from hub.conversation_members m where m.conversation_id=$2 and m.account_id=p.account_id and m.left_at is null)`, *a.Person, id).Scan(&name); err != nil {
					return httpx.Bad("That person cannot take this ticket: they are not part of its conversation. Add them first.")
				}
				args = append(args, *a.Person)
				set += fmt.Sprintf(", handler=$%d, handler_since=now(), step=greatest(step,1)", len(args))
				dest += " (" + name + ")"
				if _, err := tx.Exec(ctx, `insert into hub.ticket_assignees (org_id, ticket_id, person_id) values ($1,$2,$3) on conflict do nothing`, p.Org, id, *a.Person); err != nil {
					return err
				}
				if err := s.A.Sign.Seal(ctx, tx, "ticket_assignees", "ticket_id=$1", id); err != nil {
					return err
				}
				if err := nb.Add(ctx, tx, p.Org, *a.Person, "ticket", cur.Ref+" was escalated to you.", "#FF9B8A", cur.Ref); err != nil {
					return err
				}
			}
			text = "Escalated to " + dest
			if a.Why != "" {
				text += ": " + a.Why
			}
		case "priority":
			if err := need("Assign & hand over"); err != nil {
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
		if err := nb.Add(ctx, tx, org, id, "ticket", t.Ref+" was escalated to you.", "#FF9B8A", t.Ref); err != nil {
			return err
		}
	}
	return nil
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
		rs, err := tx.Query(ctx, `select t.conversation_id, t.ref, t.handler, t.created_at, t.due_at, t.late, t.escalated, t.hold
			from hub.tickets t where t.step < 3 and t.due_at is not null and (not t.late and t.due_at < now() or (t.due_at - t.created_at) * ($1::numeric / 100) + t.created_at < now() and t.due_at >= now())`, cfg.NearPct)
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
		}
		var rowsSeen []row
		for rs.Next() {
			var r row
			if err := rs.Scan(&r.id, &r.ref, &r.handler, &r.created, &r.due, &r.late, &r.esc, &r.hold); err != nil {
				rs.Close()
				return err
			}
			rowsSeen = append(rowsSeen, r)
		}
		rs.Close()
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
				if err := s.event(ctx, tx, org, r.id, nil, "event", "Passed its TAT and was escalated to "+cfg.Escalate); err != nil {
					return err
				}
				t := &Ticket{Ref: r.ref}
				if err := s.notifyEscalation(ctx, tx, &nb, org, t, cfg); err != nil {
					return err
				}
				if r.handler != nil && cfg.Notify.Overdue {
					if err := nb.Add(ctx, tx, org, *r.handler, "ticket", r.ref+" is overdue.", "#FF9B8A", r.ref); err != nil {
						return err
					}
				}
				late++
			} else if r.handler != nil && cfg.Notify.Nearing {
				var already bool
				if err := tx.QueryRow(ctx, `select exists (select 1 from hub.notifications where person_id=$1 and kind='ticket-near' and ref=$2)`, *r.handler, r.ref).Scan(&already); err != nil {
					return err
				}
				if !already {
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

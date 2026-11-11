// Package tickets is the ticket desk: routing, timing (TAT), hand-over and escalation. The subject, description and comments
// of a ticket are kept by the server, encrypted at rest (see text.go); its attachments stay end-to-end encrypted in its conversation.
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
	Kind         string      `json:"kind"` // "request" for a business request, "ticket" otherwise
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
	Participant  *bool       `json:"participant,omitempty"` // whether I am in this ticket's conversation (may read its text)
	Title        string      `json:"title,omitempty"`
	Description  string      `json:"description,omitempty"`
	Comments     []Comment   `json:"comments,omitempty"`
	// what the person asking may do with it (single-ticket reads): assign or reassign (head of the receiving department), escalate
	// (that head, or the person working on it)
	CanAssign   bool `json:"can_assign"`
	CanEscalate bool `json:"can_escalate"`
	// CanWork: the caller may resolve it and move it along (the person on it, the head of the place it is with, or the Solve right)
	CanWork bool `json:"can_work"`
	// Acks: who at the receiving place acknowledged a business request or communication, and when; CanAck: the caller may
	Acks   []Ack `json:"acknowledgements,omitempty"`
	CanAck bool  `json:"can_acknowledge,omitempty"`
	// ElapsedHours is how much of the TAT's running clock has been used (pauses and off-hours excluded); above TATHours means overdue.
	ElapsedHours float64    `json:"elapsed_hours"`
	ClosedAt     *time.Time `json:"closed_at,omitempty"`
	Rating       *int       `json:"rating,omitempty"` // 1 to 5 stars, given when it was closed
	AutoClosed   bool       `json:"auto_closed"`
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
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	Category     string     `json:"category"`
	Priority     string     `json:"priority"`
	ToType       string     `json:"to_type"`
	ToRef        *uuid.UUID `json:"to_ref"`
	Sync         string     `json:"sync"`
	Kind         string     `json:"kind"` // "request" for a business request (to a department only); empty for a ticket
	// FromType/FromRef: the place it is sent on behalf of, from that place's dashboard (a communication always is); empty when a person
	// raises it for themselves
	FromType string     `json:"from_type"`
	FromRef  *uuid.UUID `json:"from_ref"`
	// At backdates the ticket (history loading only; not settable over HTTP).
	At *time.Time `json:"-"`
	// Details is what emails may say about the ticket. It is only written into emails, never stored.
	Details *Details `json:"details"`
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
// RequestCategory holds a department's turnaround for business requests (one per department, set by its head). It is kept with
// the ticket categories but never offered as a ticket category.
const RequestCategory = "Business request"

// DefaultRequestTAT applies until a department sets its own: three working days.
const DefaultRequestTAT = 72

// CommunicationCategory holds a place's turnaround for the communications other places send it from their dashboards.
const CommunicationCategory = "Communication"

// DefaultCommunicationTAT applies until a place sets its own: two days.
const DefaultCommunicationTAT = 48

// KindCategory is the category that holds the turnaround for business requests or communications.
func KindCategory(kind string) string {
	if kind == "communication" {
		return CommunicationCategory
	}
	return RequestCategory
}

// DefaultKindTAT is the turnaround used before a place sets its own.
func DefaultKindTAT(kind string) int {
	if kind == "communication" {
		return DefaultCommunicationTAT
	}
	return DefaultRequestTAT
}

// SetRequestTAT sets how long a department takes to deal with a business request, from 1 hour to 3 months (2160 hours). Its head
// may set it, and so may anyone who can set up ticket categories.
func (s *Service) SetRequestTAT(ctx context.Context, p *tenant.Principal, dept uuid.UUID, hours int) error {
	return s.SetKindTAT(ctx, p, "request", dept, hours)
}

// SetKindTAT sets how long a place takes to deal with business requests (departments only) or communications (any place).
func (s *Service) SetKindTAT(ctx context.Context, p *tenant.Principal, kind string, dept uuid.UUID, hours int) error {
	if kind != "request" && kind != "communication" {
		return httpx.Bad("Choose business requests or communications.")
	}
	if hours < 1 || hours > 24*90 {
		return httpx.Bad("Choose a turnaround between 1 hour and 3 months (2160 hours).")
	}
	return s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		var unitKind string
		if err := tx.QueryRow(ctx, `select kind from hub.org_units where id=$1`, dept).Scan(&unitKind); err != nil || (kind == "request" && unitKind != "department") {
			return httpx.ErrNotFound
		}
		name := KindCategory(kind)
		head, err := isHead(ctx, tx, p.Person, &Ticket{ToType: unitKind, ToRef: &dept})
		if err != nil {
			return err
		}
		if !head {
			if err := authz.Require(ctx, tx, p.Person, "tickets", "Set up ticket categories"); err != nil {
				return err
			}
		}
		var id uuid.UUID
		err = tx.QueryRow(ctx, `select id from hub.ticket_categories where unit_type=$3 and unit_ref=$1 and name=$2`, dept, name, unitKind).Scan(&id)
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			id = app.NewID()
			if _, err := tx.Exec(ctx, `insert into hub.ticket_categories (org_id, id, unit_type, unit_ref, name, tat_hours, priority) values ($1,$2,$7,$3,$4,$5,$6)`, p.Org, id, dept, name, hours, PriorityOf(hours), unitKind); err != nil {
				return err
			}
		case err != nil:
			return err
		default:
			if _, err := tx.Exec(ctx, `update hub.ticket_categories set tat_hours=$2, priority=$3, `+integrity.Bump+` where id=$1`, id, hours, PriorityOf(hours)); err != nil {
				return err
			}
		}
		if err := s.A.Sign.Seal(ctx, tx, "ticket_categories", "id=$1", id); err != nil {
			return err
		}
		return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, fmt.Sprintf("Set the %s turnaround to %d hours", map[string]string{"request": "business request", "communication": "communication"}[kind], hours), "Tickets", p.IP, map[string]any{"unit": dept, "kind": kind, "hours": hours})
	})
}

// PriorityOf is the priority a turnaround time gives: within 8 hours High, within 24 hours Medium, anything longer Low.
func PriorityOf(tatHours int) string {
	switch {
	case tatHours <= 8:
		return "High"
	case tatHours <= 24:
		return "Medium"
	}
	return "Low"
}

func (s *Service) category(ctx context.Context, tx pgx.Tx, toType string, toRef *uuid.UUID, name string) (*Category, error) {
	c := &Category{}
	// a place that set its own categories takes only those; one that has none yet takes the organization's general ones
	row := tx.QueryRow(ctx, `select id, unit_type, unit_ref, name, tat_hours, priority, members from hub.ticket_categories
		where name=$1 and ((unit_type=$2 and unit_ref is not distinct from $3) or (unit_type='desk' and not exists (
		  select 1 from hub.ticket_categories o where o.unit_type=$2 and o.unit_ref is not distinct from $3 and o.unit_type <> 'desk' and o.name <> $4 and o.name <> $5)))
		order by (unit_type='desk') limit 1`, name, toType, toRef, RequestCategory, CommunicationCategory)
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
	var err error
	if in.Title, err = cleanText(in.Title, maxTitle, "The subject"); err != nil {
		return nil, err
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

// Acknowledge records that the caller, at the receiving place, received and noted a business request or communication.
func (s *Service) Acknowledge(ctx context.Context, p *tenant.Principal, id uuid.UUID) (*Ticket, error) {
	var out *Ticket
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		t, err := s.get(ctx, tx, p.Org, id)
		if err != nil {
			return err
		}
		ok, err := mayAcknowledge(ctx, tx, p.Person, t)
		if err != nil {
			return err
		}
		if !ok {
			return httpx.Err(403, "not_receiver", "Only the people of the place it was sent to can acknowledge it.")
		}
		tag, err := tx.Exec(ctx, `insert into hub.ticket_acks (org_id, ticket_id, person_id) values ($1,$2,$3) on conflict do nothing`, p.Org, id, p.Person)
		if err != nil {
			return err
		}
		if tag.RowsAffected() > 0 {
			if err := s.event(ctx, tx, p.Org, id, p, "event", "Acknowledged it"); err != nil {
				return err
			}
		}
		out, err = s.get(ctx, tx, p.Org, id)
		return err
	})
	if err == nil {
		s.changed(ctx, p, out)
	}
	return out, err
}

func (s *Service) Get(ctx context.Context, p *tenant.Principal, id uuid.UUID) (*Ticket, error) {
	var t *Ticket
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "tickets", "View"); err != nil {
			return err
		}
		var err error
		if t, err = s.get(ctx, tx, p.Org, id); err != nil {
			return err
		}
		ok, err := mayAssign(ctx, tx, p.Person, t)
		if err != nil {
			return err
		}
		t.CanAssign = ok
		t.CanEscalate = ok || (t.Handler != nil && *t.Handler == p.Person)
		head, err := isHead(ctx, tx, p.Person, t)
		if err != nil {
			return err
		}
		solve, err := authz.Can(ctx, tx, p.Person, "tickets", "Solve")
		if err != nil {
			return err
		}
		t.CanWork = t.RaisedBy != p.Person && (head || solve || (t.Handler != nil && *t.Handler == p.Person))
		if t.CanAck, err = mayAcknowledge(ctx, tx, p.Person, t); err != nil {
			return err
		}
		in, err := s.participant(ctx, tx, p, id)
		if err != nil || !in {
			t.Participant = &in
			return err
		}
		t.Participant = &in
		return s.fillText(ctx, tx, p.Org, []*Ticket{t}, true)
	})
	return t, err
}

func (s *Service) changed(ctx context.Context, p *tenant.Principal, t *Ticket) {
	facts := *t // the whole organization hears about changes: never the text
	facts.Title, facts.Description, facts.Comments = "", "", nil
	d, _ := json.Marshal(&facts)
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
	// Details is what the email about this action may say about the ticket (written into the email, never stored).
	Details *Details `json:"details,omitempty"`
	// Category re-files the ticket under one of the destination's categories; its time limit starts again from now.
	Category string `json:"category,omitempty"`
	// Comment: what was done (required to resolve) or a closing remark. Rating: 1 to 5 stars when closing (optional).
	Comment string `json:"comment,omitempty"`
	Rating  int    `json:"rating,omitempty"`
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
		raisedByMe := cur.RaisedBy == p.Person
		// the head of the place it is with now (its dashboard received it; or a place above that one) works on it as of right:
		// resolves, closes and moves it along, whatever their role's ticket rights. Not what their place sent or passed on.
		headHere, err := isHead(ctx, tx, p.Person, cur)
		if err != nil {
			return err
		}
		need := func(right string) error {
			// nobody works on a ticket they raised; they may still close it
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
		if cur.Step >= 4 {
			return httpx.Err(409, "ticket_closed", "This ticket is closed. It cannot be reopened or changed.")
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
			case 4: // closing: optional stars and remark
				if a.Rating < 0 || a.Rating > 5 {
					return httpx.Bad("Rate it from 1 to 5 stars.")
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
			if cur.Escalated && a.ToType == "" {
				return httpx.Err(409, "already_escalated", "This ticket was already escalated.")
			}
			a.Why = strings.TrimSpace(a.Why)
			if len(a.Why) > 300 {
				return httpx.Bad("Keep the reason under 300 characters.")
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
				set += fmt.Sprintf(", handler=$%d, handler_since=now(), step=greatest(step,1)", len(args))
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
		where uh.unit_id = $1 and (p.sector_id = $1 or p.branch_id = $1 or p.unit_id = $1 or p.department_id = $1)
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
				var already bool
				if err := tx.QueryRow(ctx, `select exists (select 1 from hub.notifications where person_id=$1 and kind='ticket-near' and ref=$2)`, *r.handler, r.ref).Scan(&already); err != nil {
					return err
				}
				if !already {
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
			  and (r.scope like 'Each %' or exists (select 1 from hub.unit_heads uh where uh.unit_id = up.id and uh.role_id = r.id))))`, person, *t.ToRef).Scan(&head)
	return head, err
}

// mayAssign: a head of the ticket's place, or (when they did not raise it) someone whose role holds "Assign & hand over".
func mayAssign(ctx context.Context, tx pgx.Tx, person uuid.UUID, t *Ticket) (bool, error) {
	head, err := isHead(ctx, tx, person, t)
	if err != nil || head {
		return head, err
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

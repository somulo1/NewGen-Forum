// Package tickets is the ticket desk: routing, timing (TAT), hand-over and escalation. The subject, description and comments
// of a ticket are kept by the server, encrypted at rest (see text.go); its attachments are plain files on the server (files.go).
// Tickets are not end-to-end encrypted: their conversation only lists the people on them.
package tickets

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/authz"
	"hub/internal/blob"
	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/tenant"
)

type Service struct {
	A     *app.App
	Store blob.Store // where attachments are kept (set by the files module)
}

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
	Files        []File      `json:"files,omitempty"`
	FileCount    int         `json:"file_count"` // attachments (lists show whether there is evidence without loading each ticket)
	// what the person asking may do with it (single-ticket reads): assign or reassign (head of the receiving department), escalate
	// (that head, or the person working on it)
	CanAssign   bool `json:"can_assign"`
	CanEscalate bool `json:"can_escalate"`
	// CanWork: the caller may resolve it and move it along (the person on it, the head of the place it is with, or the Solve right)
	CanWork bool `json:"can_work"`
	// CanChase: it is overdue and the caller raised it or heads the place it came from, so may send it to the receiving head
	CanChase bool `json:"can_chase"`
	// OriginRef: the dashboard it came from (see originOf), so heads on both sides can be shown
	OriginRef *uuid.UUID `json:"origin_ref,omitempty"`
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

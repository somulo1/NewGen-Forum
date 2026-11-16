package tickets

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/authz"
	"hub/internal/httpx"
	"hub/internal/realtime"
	"hub/internal/tenant"
)

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
		t.CanWork = head || (t.RaisedBy != p.Person && (solve || (t.Handler != nil && *t.Handler == p.Person)))
		if t.CanAck, err = mayAcknowledge(ctx, tx, p.Person, t); err != nil {
			return err
		}
		if t.OriginRef, err = originOf(ctx, tx, id); err != nil {
			return err
		}
		if t.CanChase, err = s.mayChase(ctx, tx, p.Person, t); err != nil {
			return err
		}
		in, err := s.participant(ctx, tx, p, id)
		if err != nil || !in {
			t.Participant = &in
			return err
		}
		t.Participant = &in
		if t.Files, err = s.files(ctx, tx, id); err != nil {
			return err
		}
		return s.fillText(ctx, tx, p.Org, []*Ticket{t}, true)
	})
	return t, err
}

func (s *Service) changed(ctx context.Context, p *tenant.Principal, t *Ticket) {
	facts := *t // the whole organization hears about changes: never the text
	facts.Title, facts.Description, facts.Comments, facts.Files = "", "", nil, nil
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

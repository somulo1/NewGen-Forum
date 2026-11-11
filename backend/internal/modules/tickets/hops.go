package tickets

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/integrity"
)

// A stay ("hop") is the time a ticket or business request spends with one place: from when that place received it to when it
// resolved it or escalated it on. Its due time is the place's own turnaround from the moment it arrived, so every place is judged
// on its own time only. While a stay is open, its due time is the ticket's (pauses on hold move it); it is copied in on leaving.

// openHop starts a stay at the place the ticket has just reached.
func (s *Service) openHop(ctx context.Context, tx pgx.Tx, org, ticket uuid.UUID, toType string, toRef *uuid.UUID, cat string, tat int, at time.Time) error {
	id := app.NewID()
	if _, err := tx.Exec(ctx, `insert into hub.ticket_hops (org_id, id, ticket_id, unit_type, unit_ref, category, tat_hours, received_at) values ($1,$2,$3,$4,$5,$6,$7,$8)`,
		org, id, ticket, toType, toRef, cat, tat, at); err != nil {
		return err
	}
	return s.A.Sign.Seal(ctx, tx, "ticket_hops", "id=$1", id)
}

// ensureHop gives a ticket made before stays were kept its one stay, from when it was raised, so it can be closed like any other.
func (s *Service) ensureHop(ctx context.Context, tx pgx.Tx, org uuid.UUID, t *Ticket) error {
	var any bool
	if err := tx.QueryRow(ctx, `select exists (select 1 from hub.ticket_hops where ticket_id=$1)`, t.ID).Scan(&any); err != nil || any {
		return err
	}
	return s.openHop(ctx, tx, org, t.ID, t.ToType, t.ToRef, t.Category, t.TATHours, t.CreatedAt)
}

// closeHop ends the open stay: resolved or escalated, on time when that happened by the stay's due time.
func (s *Service) closeHop(ctx context.Context, tx pgx.Tx, org uuid.UUID, t *Ticket, outcome string, at time.Time) error {
	if err := s.ensureHop(ctx, tx, org, t); err != nil {
		return err
	}
	var id uuid.UUID
	err := tx.QueryRow(ctx, `update hub.ticket_hops set left_at=$2, outcome=$3, due_at=$4, on_time = ($4::timestamptz is null or $2 <= $4), handler=$5, updated_at=now(), `+integrity.Bump+`
		where id = (select id from hub.ticket_hops where ticket_id=$1 and left_at is null order by received_at desc limit 1) returning id`, t.ID, at, outcome, t.DueAt, t.Handler).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	return s.A.Sign.Seal(ctx, tx, "ticket_hops", "id=$1", id)
}

// reopenHop: a resolved ticket taken back up (reopened) is open again with the place that resolved it, on that stay's clock.
func (s *Service) reopenHop(ctx context.Context, tx pgx.Tx, t *Ticket) error {
	var id uuid.UUID
	err := tx.QueryRow(ctx, `update hub.ticket_hops set left_at=null, outcome=null, due_at=null, on_time=null, handler=null, updated_at=now(), `+integrity.Bump+`
		where id = (select id from hub.ticket_hops where ticket_id=$1 order by received_at desc limit 1) and outcome='resolved' returning id`, t.ID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	return s.A.Sign.Seal(ctx, tx, "ticket_hops", "id=$1", id)
}

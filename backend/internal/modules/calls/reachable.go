package calls

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/httpx"
	"hub/internal/realtime"
	"hub/internal/tenant"
)

// busy returns which of these people are in a call right now: they joined a call that is still going and have not left it, or a call
// is ringing for them (or they are the one ringing) that was started within the ring time. A call row that never closed (a crash)
// stops counting after two hours.
func (m *Module) busy(ctx context.Context, tx pgx.Tx, people []uuid.UUID) (map[uuid.UUID]bool, error) {
	out := map[uuid.UUID]bool{}
	rs, err := tx.Query(ctx, `select distinct cp.person_id from hub.call_participants cp join hub.calls c on c.org_id = cp.org_id and c.id = cp.call_id
		where cp.person_id = any($1) and cp.left_at is null and c.ended_at is null and c.status in ('ringing','active') and c.started_at > now() - interval '2 hours'
		  and ((cp.joined_at is not null) or (c.status = 'ringing' and c.started_at > now() - interval '50 seconds'))`, people)
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	for rs.Next() {
		var id uuid.UUID
		if err := rs.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = true
	}
	return out, rs.Err()
}

// checkReachable refuses a private call, before anything is created or rung, when the caller is busy, or the other person is offline
// or busy. The answers say which.
func (m *Module) checkReachable(ctx context.Context, tx pgx.Tx, p *tenant.Principal, invited []uuid.UUID) error {
	if len(invited) == 0 {
		return nil
	}
	b, err := m.busy(ctx, tx, append([]uuid.UUID{p.Person}, invited...))
	if err != nil {
		return err
	}
	if b[p.Person] {
		return httpx.Err(409, "caller_busy", "You are already in a call or have one ringing. Finish or answer it first.")
	}
	for _, id := range invited {
		var name string
		_ = tx.QueryRow(ctx, `select first_name || ' ' || last_name from hub.people where id=$1`, id).Scan(&name)
		name = strings.TrimSpace(name)
		if name == "" {
			name = "That person"
		}
		// online means their app holds a live connection to Hub (it refreshes this marker every 25 seconds)
		if !realtime.Online(ctx, m.A.Rdb, realtime.PresenceKey(p.Org, id.String())) {
			return httpx.Err(409, "callee_offline", name+" is offline right now. Try again when they are online.")
		}
		if b[id] {
			return httpx.Err(409, "callee_busy", name+" is on another call right now. Try again in a moment.")
		}
	}
	return nil
}

package tickets

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/authz"
	"hub/internal/httpx"
	"hub/internal/tenant"
)

// Intake: a place that receives tickets only by escalation, and only from one chosen place (see migration 0092).
type Intake struct {
	Unit     uuid.UUID `json:"unit"`
	OnlyFrom uuid.UUID `json:"only_from"`
}

func (s *Service) Intakes(ctx context.Context, p *tenant.Principal) ([]Intake, error) {
	out := []Intake{}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		rs, err := tx.Query(ctx, `select unit_id, only_from from hub.ticket_intake`)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var x Intake
			if err := rs.Scan(&x.Unit, &x.OnlyFrom); err != nil {
				return err
			}
			out = append(out, x)
		}
		return rs.Err()
	})
	return out, err
}

// SetIntake: the head of the receiving place (or someone who sets up ticket categories) chooses the places it takes escalations
// from; an empty list opens it to everyone again.
func (s *Service) SetIntake(ctx context.Context, p *tenant.Principal, unit uuid.UUID, onlyFrom []uuid.UUID) error {
	if len(onlyFrom) > 200 {
		return httpx.Bad("Choose fewer places.")
	}
	return s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		var kind string
		if err := tx.QueryRow(ctx, `select kind from hub.org_units where id=$1`, unit).Scan(&kind); err != nil {
			return httpx.ErrNotFound
		}
		head, err := isHead(ctx, tx, p.Person, &Ticket{ToType: kind, ToRef: &unit})
		if err != nil {
			return err
		}
		if !head {
			if err := authz.Require(ctx, tx, p.Person, "tickets", "Set up ticket categories"); err != nil {
				return err
			}
		}
		for _, f := range onlyFrom {
			if f == unit {
				return httpx.Bad("A place cannot escalate to itself.")
			}
		}
		var n int
		if err := tx.QueryRow(ctx, `select count(*) from hub.org_units where id = any($1)`, onlyFrom).Scan(&n); err != nil {
			return err
		}
		if n != len(dedupe(onlyFrom)) {
			return httpx.Bad("One of those places does not exist.")
		}
		if _, err := tx.Exec(ctx, `delete from hub.ticket_intake where unit_id=$1`, unit); err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `insert into hub.ticket_intake (org_id, unit_id, only_from) select $1, $2, x from unnest($3::uuid[]) x on conflict do nothing`, p.Org, unit, onlyFrom)
		return err
	})
}

func dedupe(in []uuid.UUID) []uuid.UUID {
	seen := map[uuid.UUID]bool{}
	var out []uuid.UUID
	for _, x := range in {
		if !seen[x] {
			seen[x] = true
			out = append(out, x)
		}
	}
	return out
}

// checkIntakeEscalation: a place closed to everyone else takes escalations only from members of the places it chose.
func checkIntakeEscalation(ctx context.Context, tx pgx.Tx, person, to uuid.UUID, toName string) error {
	var closed, member bool
	if err := tx.QueryRow(ctx, `select exists (select 1 from hub.ticket_intake where unit_id=$2),
			exists (select 1 from hub.ticket_intake i join hub.people pe on pe.id=$1 where i.unit_id=$2 and i.only_from in (pe.sector_id, pe.branch_id, pe.unit_id, pe.department_id))`, person, to).Scan(&closed, &member); err != nil {
		return err
	}
	if closed && !member {
		return httpx.Err(403, "intake_closed", toName+" only takes tickets escalated from the places it chose.")
	}
	return nil
}

// intakeClosed: whether a place takes tickets only by escalation (from places it chose).
func intakeClosed(ctx context.Context, tx pgx.Tx, unit uuid.UUID) (bool, error) {
	var closed bool
	err := tx.QueryRow(ctx, `select exists (select 1 from hub.ticket_intake where unit_id=$1)`, unit).Scan(&closed)
	return closed, err
}

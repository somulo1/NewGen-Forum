package tickets

import (
	"context"
	"errors"

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

// SetIntake: the head of the receiving place (or someone who sets up ticket categories) chooses; onlyFrom nil opens it again.
func (s *Service) SetIntake(ctx context.Context, p *tenant.Principal, unit uuid.UUID, onlyFrom *uuid.UUID) error {
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
		if onlyFrom == nil {
			_, err := tx.Exec(ctx, `delete from hub.ticket_intake where unit_id=$1`, unit)
			return err
		}
		if *onlyFrom == unit {
			return httpx.Bad("Choose another place: a place cannot escalate to itself.")
		}
		var ok bool
		if err := tx.QueryRow(ctx, `select exists (select 1 from hub.org_units where id=$1)`, *onlyFrom).Scan(&ok); err != nil || !ok {
			return httpx.Bad("That place does not exist.")
		}
		_, err = tx.Exec(ctx, `insert into hub.ticket_intake (org_id, unit_id, only_from) values ($1,$2,$3)
			on conflict (org_id, unit_id) do update set only_from=excluded.only_from, updated_at=now()`, p.Org, unit, *onlyFrom)
		return err
	})
}

// intakeOf returns the only place a unit takes escalations from, or nil when anyone may send to it.
func intakeOf(ctx context.Context, tx pgx.Tx, unit uuid.UUID) (*uuid.UUID, string, error) {
	var from uuid.UUID
	var name string
	err := tx.QueryRow(ctx, `select i.only_from, u.name from hub.ticket_intake i join hub.org_units u on u.id = i.only_from where i.unit_id=$1`, unit).Scan(&from, &name)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, "", nil
	}
	if err != nil {
		return nil, "", err
	}
	return &from, name, nil
}

// checkIntakeEscalation: only a member of the chosen place may escalate into a place that is closed to everyone else.
func checkIntakeEscalation(ctx context.Context, tx pgx.Tx, person, to uuid.UUID, toName string) error {
	from, fromName, err := intakeOf(ctx, tx, to)
	if err != nil || from == nil {
		return err
	}
	var member bool
	if err := tx.QueryRow(ctx, `select coalesce($2 in (sector_id, branch_id, unit_id, department_id), false) from hub.people where id=$1`, person, *from).Scan(&member); err != nil {
		return err
	}
	if !member {
		return httpx.Err(403, "intake_closed", toName+" only takes tickets escalated from "+fromName+".")
	}
	return nil
}

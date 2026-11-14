package org

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/authz"
	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/tenant"
)

// SetHeadPerson chooses who leads a place: someone who holds the place's head role, or one of its members. They are placed in it
// (and in every place above it) so the head is always a member. nil clears the choice: the head role decides again.
func (s *Service) SetHeadPerson(ctx context.Context, p *tenant.Principal, unit uuid.UUID, person *uuid.UUID) error {
	return s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "settings", "Edit organization profile"); err != nil {
			return err
		}
		var name string
		if err := tx.QueryRow(ctx, `select name from hub.org_units where id=$1`, unit).Scan(&name); err != nil {
			return httpx.ErrNotFound
		}
		if person == nil {
			if _, err := tx.Exec(ctx, `delete from hub.unit_head_people where unit_id=$1`, unit); err != nil {
				return err
			}
			return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Cleared the chosen head of "+name, "Settings", p.IP, map[string]any{"unit": unit})
		}
		var ok bool
		var who string
		if err := tx.QueryRow(ctx, `select first_name || ' ' || last_name,
				$2 in (sector_id, branch_id, unit_id, department_id)
				or exists (select 1 from hub.person_roles pr join hub.unit_heads uh on uh.role_id = pr.role_id where pr.person_id = pe.id and uh.unit_id = $2)
			from hub.people pe where id=$1 and status = 'active'`, *person, unit).Scan(&who, &ok); err != nil {
			return httpx.Bad("Choose an active person.")
		}
		if !ok {
			return httpx.Bad(who + " neither holds the role that heads " + name + " nor is a member of it.")
		}
		// the head is a member: placed here and in every place above it
		if _, err := tx.Exec(ctx, `with recursive up(id, kind, parent_id) as (select id, kind, parent_id from hub.org_units where id=$2
				union all select u.id, u.kind, u.parent_id from hub.org_units u join up on u.id = up.parent_id)
			update hub.people set
				sector_id = coalesce((select id from up where kind='sector' limit 1), sector_id),
				branch_id = coalesce((select id from up where kind='branch' limit 1), branch_id),
				unit_id = coalesce((select id from up where kind='unit' limit 1), unit_id),
				department_id = coalesce((select id from up where kind='department' limit 1), department_id),
				updated_at = now(), `+integrity.Bump+` where id=$1`, *person, unit); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "people", "id=$1", *person); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `insert into hub.unit_head_people (org_id, unit_id, person_id) values ($1,$2,$3)
			on conflict (org_id, unit_id) do update set person_id=excluded.person_id, updated_at=now()`, p.Org, unit, *person); err != nil {
			return err
		}
		return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Made "+who+" the head of "+name, "Settings", p.IP, map[string]any{"unit": unit, "person": *person})
	})
}

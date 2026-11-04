package org

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/authz"
	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/tenant"
)

type Unit struct {
	ID     uuid.UUID  `json:"id"`
	Kind   string     `json:"kind"`
	Parent *uuid.UUID `json:"parent_id"`
	Name   string     `json:"name"`
	Code   string     `json:"code"`
	// Target is the performance threshold (percent); sent with the structure so every dashboard card can colour itself.
	Target int `json:"target,omitempty"`
	// HeadRole is the role that heads this unit (its key, and its name for display); Heads are the people who hold that role and are
	// placed here.
	HeadRole     string `json:"head_role,omitempty"`
	HeadRoleName string `json:"head_role_name,omitempty"`
	Heads        []Head `json:"heads,omitempty"`
}

type Head struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// parentKinds says what a unit of each kind may sit under. Every level is optional (an organization may have only
// departments, or sectors and units but no branches), so a unit may also stand at the top level or under the nearest level
// the organization does have.
var parentKinds = map[string][]string{
	"sector":     nil,
	"branch":     {"sector"},
	"unit":       {"branch", "sector"},
	"department": {"branch", "sector", "unit"},
}

func (s *Service) Structure(ctx context.Context, p *tenant.Principal) ([]Unit, error) {
	out := []Unit{}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		rs, err := tx.Query(ctx, `select u.id, u.kind, u.parent_id, u.name, coalesce(u.code,''), coalesce(t.target_pct, 0), coalesce(r.key,''), coalesce(r.name,'')
			from hub.org_units u left join hub.unit_targets t on t.unit_id = u.id
			left join hub.unit_heads uh on uh.unit_id = u.id left join hub.roles r on r.id = uh.role_id
			order by u.kind, u.name`)
		if err != nil {
			return err
		}
		defer rs.Close()
		idx := map[uuid.UUID]int{}
		for rs.Next() {
			var u Unit
			if err := rs.Scan(&u.ID, &u.Kind, &u.Parent, &u.Name, &u.Code, &u.Target, &u.HeadRole, &u.HeadRoleName); err != nil {
				return err
			}
			idx[u.ID] = len(out)
			out = append(out, u)
		}
		if err := rs.Err(); err != nil {
			return err
		}
		rs.Close()
		// the heads: people who hold the unit's head role and are placed in it
		hs, err := tx.Query(ctx, `select uh.unit_id, p.id, p.first_name || ' ' || p.last_name
			from hub.unit_heads uh join hub.person_roles pr on pr.role_id = uh.role_id
			join hub.people p on p.id = pr.person_id and p.status <> 'deleted'
			where p.sector_id = uh.unit_id or p.branch_id = uh.unit_id or p.unit_id = uh.unit_id or p.department_id = uh.unit_id
			order by p.first_name, p.last_name`)
		if err != nil {
			return err
		}
		defer hs.Close()
		for hs.Next() {
			var unit uuid.UUID
			var h Head
			if err := hs.Scan(&unit, &h.ID, &h.Name); err != nil {
				return err
			}
			if i, ok := idx[unit]; ok {
				out[i].Heads = append(out[i].Heads, h)
			}
		}
		return hs.Err()
	})
	return out, err
}

// SetHead gives a unit its head role (the role's key), or takes it away (an empty key). Everyone who holds that role and is placed in
// the unit then heads it, with the rights the role gives them.
func (s *Service) SetHead(ctx context.Context, p *tenant.Principal, unit uuid.UUID, roleKey string) error {
	return s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "settings", "Edit organization profile"); err != nil {
			return err
		}
		var name, kind string
		if err := tx.QueryRow(ctx, `select name, kind from hub.org_units where id=$1 for update`, unit).Scan(&name, &kind); err != nil {
			return httpx.ErrNotFound
		}
		if roleKey == "" {
			if _, err := tx.Exec(ctx, `delete from hub.unit_heads where unit_id=$1`, unit); err != nil {
				return err
			}
			return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Removed the head role of "+name, "Settings", p.IP, map[string]any{"unit": unit})
		}
		var role uuid.UUID
		var roleName string
		if err := tx.QueryRow(ctx, `select id, name from hub.roles where key=$1`, roleKey).Scan(&role, &roleName); err != nil {
			return httpx.Bad("Choose a role from the list.")
		}
		if _, err := tx.Exec(ctx, `insert into hub.unit_heads (org_id, unit_id, role_id) values ($1,$2,$3) on conflict do nothing`, p.Org, unit, role); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "unit_heads", "unit_id=$1", unit); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `update hub.unit_heads set role_id=$2, updated_at=now(), `+integrity.Bump+` where unit_id=$1`, unit, role); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "unit_heads", "unit_id=$1", unit); err != nil {
			return err
		}
		return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Made the "+roleName+" role the head of "+name, "Settings", p.IP, map[string]any{"unit": unit, "role": roleKey})
	})
}

func (s *Service) SaveUnit(ctx context.Context, p *tenant.Principal, id *uuid.UUID, in Unit) (*Unit, error) {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" || len(in.Name) > 80 {
		return nil, httpx.Bad("Enter a name.")
	}
	var out Unit
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "settings", "Edit organization profile"); err != nil {
			return err
		}
		if id == nil {
			lv, err := s.levels(ctx, tx, p)
			if err != nil {
				return err
			}
			pk := ""
			if in.Parent != nil {
				if err := tx.QueryRow(ctx, `select kind from hub.org_units where id=$1`, *in.Parent).Scan(&pk); err != nil {
					return httpx.Bad("That parent does not exist.")
				}
			}
			if err := lv.checkPlace(in.Kind, pk); err != nil {
				return err
			}
			out = Unit{ID: uuidNew(), Kind: in.Kind, Parent: in.Parent, Name: in.Name, Code: in.Code}
			if _, err := tx.Exec(ctx, `insert into hub.org_units (id, org_id, kind, parent_id, name, code) values ($1,$2,$3,$4,$5,nullif($6,''))`, out.ID, p.Org, out.Kind, out.Parent, out.Name, out.Code); err != nil {
				if strings.Contains(err.Error(), "org_units_org_id_kind_name_key") {
					return httpx.Err(409, "duplicate", "There is already a "+in.Kind+" with that name.")
				}
				return err
			}
			if err := s.A.Sign.Seal(ctx, tx, "org_units", "id=$1", out.ID); err != nil {
				return err
			}
			return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Added "+in.Kind+" "+in.Name, "Settings", p.IP, out)
		}
		tag, err := tx.Exec(ctx, `update hub.org_units set name=$2, code=nullif($3,''), updated_at=now(), `+integrity.Bump+` where id=$1`, *id, in.Name, in.Code)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return httpx.ErrNotFound
		}
		if err := s.A.Sign.Seal(ctx, tx, "org_units", "id=$1", *id); err != nil {
			return err
		}
		out = Unit{ID: *id, Name: in.Name, Code: in.Code}
		return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Renamed a unit to "+in.Name, "Settings", p.IP, out)
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteUnit removes a sector, branch, unit or department that nothing depends on: no units under it, nobody placed in
// it, no ticket sent to it. Its own ticket categories and threshold go with it.
func (s *Service) DeleteUnit(ctx context.Context, p *tenant.Principal, id uuid.UUID) error {
	return s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "settings", "Edit organization profile"); err != nil {
			return err
		}
		var name, kind string
		if err := tx.QueryRow(ctx, `select name, kind from hub.org_units where id=$1 for update`, id).Scan(&name, &kind); err != nil {
			return httpx.ErrNotFound
		}
		var kids, people, tickets int
		if err := tx.QueryRow(ctx, `select (select count(*) from hub.org_units where parent_id=$1),
			(select count(*) from hub.people where status <> 'deleted' and (sector_id=$1 or branch_id=$1 or unit_id=$1 or department_id=$1)),
			(select count(*) from hub.tickets where to_ref=$1 and step < 4)`, id).Scan(&kids, &people, &tickets); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `delete from hub.unit_heads where unit_id=$1`, id); err != nil {
			return err
		}
		switch {
		case kids > 0:
			return httpx.Err(409, "in_use", fmt.Sprintf("%s has %d unit(s) under it. Move or remove those first.", name, kids))
		case people > 0:
			return httpx.Err(409, "in_use", fmt.Sprintf("%d people are placed in %s. Move them to another place first.", people, name))
		case tickets > 0:
			return httpx.Err(409, "in_use", fmt.Sprintf("%d open tickets are sent to %s.", tickets, name))
		}
		for _, q := range []string{`delete from hub.unit_targets where unit_id=$1`, `delete from hub.ticket_categories where unit_ref=$1`, `delete from hub.org_units where id=$1`} {
			if _, err := tx.Exec(ctx, q, id); err != nil {
				return err
			}
		}
		return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Removed "+kind+" "+name, "Settings", p.IP, map[string]any{"unit": id})
	})
}

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
		rs, err := tx.Query(ctx, `select u.id, u.kind, u.parent_id, u.name, coalesce(u.code,''), coalesce(t.target_pct, 0) from hub.org_units u left join hub.unit_targets t on t.unit_id = u.id order by u.kind, u.name`)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var u Unit
			if err := rs.Scan(&u.ID, &u.Kind, &u.Parent, &u.Name, &u.Code, &u.Target); err != nil {
				return err
			}
			out = append(out, u)
		}
		return rs.Err()
	})
	return out, err
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
			allowed, ok := parentKinds[in.Kind]
			if !ok {
				return httpx.Bad("Unknown kind of unit.")
			}
			if in.Parent != nil {
				var pk string
				if err := tx.QueryRow(ctx, `select kind from hub.org_units where id=$1`, *in.Parent).Scan(&pk); err != nil {
					return httpx.Bad("That parent does not exist.")
				}
				ok := false
				for _, k := range allowed {
					ok = ok || k == pk
				}
				if !ok {
					return httpx.Bad("A " + in.Kind + " cannot sit under a " + pk + ".")
				}
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

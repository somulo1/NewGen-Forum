package org

import (
	"context"
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

// parentKinds says what a unit of each kind may sit under (nil = top level, or any of the listed kinds).
var parentKinds = map[string][]string{
	"sector":     nil,
	"branch":     {"sector"},
	"unit":       {"branch"},
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
			if in.Parent == nil && allowed != nil && in.Kind != "department" {
				return httpx.Bad("Choose where it belongs.")
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

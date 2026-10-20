package access

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Pack is a set of roles with what each may do, and optionally approval levels. The starter pack (starter.go) is what a real
// organization can apply; tests build their own in internal/testkit.
type Pack struct {
	Roles    []Role                         `json:"roles"`
	Defaults map[string]map[string][]string `json:"defaults"` // role -> module -> allowed rights
	Flows    map[string]DemoFlow            `json:"flows"`    // flow key -> levels and the role that approves each level
}

type DemoFlow struct {
	Levels    int      `json:"levels"`
	Approvers []string `json:"approvers"` // role names; a name with no role is skipped (the owner can always stand in)
}

// InstallPack writes a set of roles, what each may do, approval levels and a default role into an organization. It is safe to
// run again: a role that already exists is left exactly as it is (an organization's own changes are never overwritten), and
// only roles created by this call get their rights, approval levels and default. It returns the new role ids by key.
func InstallPack(ctx context.Context, tx pgx.Tx, org uuid.UUID, d *Pack, newID func() uuid.UUID, seal func(table, where string, args ...any) error) (map[string]uuid.UUID, error) {
	ids := map[string]uuid.UUID{}
	for _, r := range d.Roles {
		if r.Locked {
			continue // the owner role already exists
		}
		var have bool
		if err := tx.QueryRow(ctx, `select exists (select 1 from hub.roles where key=$1)`, r.ID).Scan(&have); err != nil {
			return nil, err
		}
		if have {
			continue
		}
		id := newID()
		ids[r.ID] = id
		scope := r.Scope
		if scope == "" {
			scope = "Whole organization"
		}
		if _, err := tx.Exec(ctx, `insert into hub.roles (id, org_id, key, name, description, scope, locked, needs_level) values ($1,$2,$3,$4,$5,$6,false,nullif($7,''))`, id, org, r.ID, r.Name, r.Desc, scope, r.Needs); err != nil {
			return nil, err
		}
	}
	if err := seal("roles", "org_id=$1", org); err != nil {
		return nil, err
	}
	for role, mods := range d.Defaults {
		id, ok := ids[role]
		if !ok {
			continue
		}
		for mod, rights := range mods {
			for _, r := range rights {
				if _, err := tx.Exec(ctx, `insert into hub.role_rights (org_id, role_id, module, right_name, allowed) values ($1,$2,$3,$4,true) on conflict do nothing`, org, id, mod, r); err != nil {
					return nil, err
				}
			}
		}
	}
	if err := seal("role_rights", "org_id=$1", org); err != nil {
		return nil, err
	}
	// approval levels, and which sample role may decide each one
	for key, f := range d.Flows {
		var module string
		if err := tx.QueryRow(ctx, `select module from hub.approval_flows where key=$1`, key).Scan(&module); err != nil {
			continue
		}
		if f.Levels == 0 && len(f.Approvers) == 0 {
			continue
		}
		if _, err := tx.Exec(ctx, `update hub.approval_flows set levels=$2, updated_at=now(), ver = ver + 1, mac = hub.pending_mac() where key=$1`, key, f.Levels); err != nil {
			return nil, err
		}
		for lvl, who := range f.Approvers {
			if lvl >= f.Levels {
				break
			}
			for _, r := range d.Roles {
				if id, ok := ids[r.ID]; ok && strings.EqualFold(r.Name, who) {
					if _, err := tx.Exec(ctx, `insert into hub.role_rights (org_id, role_id, module, right_name, allowed) values ($1,$2,$3,$4,true) on conflict (org_id, role_id, module, right_name) do update set allowed=true`, org, id, module, ApproveRight(lvl+1)); err != nil {
						return nil, err
					}
				}
			}
		}
	}
	if err := seal("approval_flows", "org_id=$1", org); err != nil {
		return nil, err
	}
	if err := seal("role_rights", "org_id=$1", org); err != nil {
		return nil, err
	}
	if id, ok := ids["staff"]; ok {
		if _, err := tx.Exec(ctx, `insert into hub.default_role (org_id, role_id) values ($1,$2) on conflict do nothing`, org, id); err != nil {
			return nil, err
		}
		if err := seal("default_role", "org_id=$1", org); err != nil {
			return nil, err
		}
	}
	return ids, nil
}

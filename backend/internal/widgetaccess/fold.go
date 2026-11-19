package widgetaccess

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/integrity"
	"hub/internal/orgsettings"
)

var dashboards = []string{"departments", "sectors", "branches", "units"}

// FoldOrg retires an organization's user groups. Every right a user group gave is written onto the roles in it, so nobody loses
// what they could do, and the groups are removed. A dashboard that was opened by user groups only keeps exactly the roles the
// groups let in. Widget settings lose their old audience fields (who sees a widget is now its right on Roles & rights). It does
// nothing once the organization has been folded, and says whether it changed anything.
func FoldOrg(ctx context.Context, a *app.App, tx pgx.Tx, org uuid.UUID) (bool, error) {
	if _, err := tx.Exec(ctx, `select pg_advisory_xact_lock(hashtext('fold-user-groups:' || $1))`, org.String()); err != nil {
		return false, err
	}
	raw := map[string]map[string]any{}
	if err := orgsettings.Get(ctx, a, tx, org, "widgets", &raw); err != nil {
		return false, err
	}
	stale := false
	for _, c := range raw {
		for _, k := range []string{"see", "groups", "access", "managerGroups", "managersOff", "manage"} {
			if _, ok := c[k]; ok {
				stale = true
			}
		}
	}
	var groups bool
	if err := tx.QueryRow(ctx, `select exists (select 1 from hub.role_groups)`).Scan(&groups); err != nil || (!groups && !stale) {
		return false, err
	}
	if groups {
		if err := foldRights(ctx, a, tx, org, raw); err != nil {
			return false, err
		}
	}
	if stale {
		all := map[string]Config{}
		if err := orgsettings.Get(ctx, a, tx, org, "widgets", &all); err != nil {
			return false, err
		}
		if err := orgsettings.Set(ctx, a, tx, org, "widgets", all); err != nil {
			return false, err
		}
	}
	return true, a.Audit(ctx, tx, org, nil, "Hub", "Retired user groups: their permissions now sit on the roles that were in them", "Roles", "", nil)
}

func foldRights(ctx context.Context, a *app.App, tx pgx.Tx, org uuid.UUID, raw map[string]map[string]any) error {
	// every right a group gave (a dashboard's "View dashboard" follows its own rule below)
	if _, err := tx.Exec(ctx, `insert into hub.role_rights (org_id, role_id, module, right_name, allowed)
		select distinct $1::uuid, gr.role_id, g.module, g.right_name, true from hub.role_group_rights g join hub.role_group_roles gr on gr.group_id = g.group_id
		where g.allowed and not (g.module = any($2) and g.right_name = 'View dashboard')
		on conflict (org_id, role_id, module, right_name) do update set allowed = true, `+integrity.BumpOn("role_rights")+` where not hub.role_rights.allowed`, org, dashboards); err != nil {
		return err
	}
	for _, m := range dashboards {
		c := raw[m]
		if c["access"] != "groups" && !(c["access"] == nil && c["see"] == "Specific groups") {
			continue // opened by roles: groups never counted for it
		}
		if _, err := tx.Exec(ctx, `insert into hub.role_rights (org_id, role_id, module, right_name, allowed)
			select $1::uuid, r.id, $2::text, 'View dashboard', exists (select 1 from hub.role_group_roles gr join hub.role_group_rights g on g.group_id = gr.group_id
				where gr.role_id = r.id and g.module = $2 and g.right_name = 'View dashboard' and g.allowed)
			from hub.roles r where not r.locked
			on conflict (org_id, role_id, module, right_name) do update set allowed = excluded.allowed, `+integrity.BumpOn("role_rights")+`
			where hub.role_rights.allowed <> excluded.allowed`, org, m); err != nil {
			return err
		}
	}
	if err := a.Sign.Seal(ctx, tx, "role_rights", "org_id=$1", org); err != nil {
		return err
	}
	for _, q := range []string{`delete from hub.role_group_rights`, `delete from hub.role_group_roles`, `delete from hub.role_groups`} {
		if _, err := tx.Exec(ctx, q); err != nil {
			return err
		}
	}
	return nil
}

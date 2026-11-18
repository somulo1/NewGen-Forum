// Package widgetaccess decides, on the server, which widgets (and so which dashboards) a person may use. A widget the
// organization switched on is for everyone, or only for the groups chosen for it in Widget settings: people placed in a
// department, sector, branch or unit; members of a forum; or a user group (every role in it).
//
// The four dashboards (departments, sectors, branches, units) follow one rule, set in that dashboard's Widget settings (see
// Dashboard): the sidebar's My department / My sector / My branch / My unit, the dashboard itself (even at a typed address) and
// sending tickets on a place's behalf all ask the same question, so they can never disagree.
package widgetaccess

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/orgsettings"
	"hub/internal/tenant"
)

type cfg struct {
	Enabled bool     `json:"enabled"`
	See     string   `json:"see"`
	Groups  []string `json:"groups"`
	Access  string   `json:"access"` // dashboards: "roles" or "groups"
}

// Dashboard access modes, chosen in the dashboard's Widget settings.
const (
	ByRoles  = "roles"  // the roles given "View dashboard" directly open it
	ByGroups = "groups" // only the roles in a user group given "View dashboard" open it; a role's own setting does not count
)

// ModeOf is the dashboard's access mode. A dashboard saved before the choice existed, limited to "Specific groups", was
// already a user-group dashboard.
func ModeOf(access, see string) string {
	if access == ByGroups || (access == "" && see == "Specific groups") {
		return ByGroups
	}
	return ByRoles
}

// IsDashboard says whether a widget id is one of the four place dashboards.
func IsDashboard(id string) bool { return placeCol[id] != "" }

// dashboardRule applies the one rule for a dashboard level: switched on, then a super admin always; a personal grant or denial
// on the person (People page) is their own exception; otherwise the mode decides between the role's own "View dashboard" and
// that of a user group the role is in.
func dashboardRule(ctx context.Context, tx pgx.Tx, person uuid.UUID, module string, c cfg) (bool, error) {
	if !c.Enabled {
		return false, nil
	}
	var locked, direct, viaGroup bool
	var personal *bool
	err := tx.QueryRow(ctx, `select coalesce(r.locked,false),
			coalesce((select rr.allowed from hub.role_rights rr where rr.role_id = pr.role_id and rr.module=$2 and rr.right_name='View dashboard'), false),
			exists (select 1 from hub.role_group_roles gr join hub.role_group_rights g on g.group_id = gr.group_id
			        where gr.role_id = pr.role_id and g.module=$2 and g.right_name='View dashboard' and g.allowed),
			(select x.allowed from hub.person_rights x where x.person_id = p.id and x.module=$2 and x.right_name='View dashboard')
		from hub.people p left join hub.person_roles pr on pr.person_id = p.id left join hub.roles r on r.id = pr.role_id
		where p.id=$1 and p.status='active'`, person, module).Scan(&locked, &direct, &viaGroup, &personal)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	switch {
	case err != nil:
		return false, err
	case locked:
		return true, nil
	case personal != nil:
		return *personal, nil
	case ModeOf(c.Access, c.See) == ByGroups:
		return viaGroup, nil
	}
	return direct, nil
}

// Dashboard says whether the person may open the dashboards of one level (departments, sectors, branches, units).
func Dashboard(ctx context.Context, a *app.App, tx pgx.Tx, p *tenant.Principal, module string) (bool, error) {
	all := map[string]cfg{}
	if err := orgsettings.Get(ctx, a, tx, p.Org, "widgets", &all); err != nil {
		return false, err
	}
	return dashboardRule(ctx, tx, p.Person, module, all[module])
}

var placeCol = map[string]string{"departments": "department_id", "sectors": "sector_id", "branches": "branch_id", "units": "unit_id"}

// Allowed lists the widget ids the person may use (switched on, and they are in its audience).
func Allowed(ctx context.Context, a *app.App, tx pgx.Tx, p *tenant.Principal) (map[string]bool, error) {
	all := map[string]cfg{}
	if err := orgsettings.Get(ctx, a, tx, p.Org, "widgets", &all); err != nil {
		return nil, err
	}
	// what this person belongs to, in one go
	var place = map[string]string{}
	var sec, br, un, de *uuid.UUID
	var role *uuid.UUID
	if err := tx.QueryRow(ctx, `select p.sector_id, p.branch_id, p.unit_id, p.department_id, pr.role_id from hub.people p
		left join hub.person_roles pr on pr.person_id = p.id where p.id=$1`, p.Person).Scan(&sec, &br, &un, &de, &role); err != nil {
		return nil, err
	}
	for k, v := range map[string]*uuid.UUID{"sectors": sec, "branches": br, "units": un, "departments": de} {
		if v != nil {
			place[k] = v.String()
		}
	}
	groups := map[string]bool{}
	if role != nil {
		// user groups are named by their key (as Widget settings stores them)
		rs, err := tx.Query(ctx, `select g.key from hub.role_group_roles rr join hub.role_groups g on g.id = rr.group_id where rr.role_id=$1`, *role)
		if err != nil {
			return nil, err
		}
		for rs.Next() {
			var g string
			if rs.Scan(&g) == nil {
				groups[g] = true
			}
		}
		rs.Close()
	}
	forums := map[string]bool{}
	rs, err := tx.Query(ctx, `select conversation_id from hub.conversation_members m join hub.conversations c on c.id = m.conversation_id and c.kind='forum'
		where m.account_id=$1 and m.left_at is null`, p.Account)
	if err != nil {
		return nil, err
	}
	for rs.Next() {
		var c uuid.UUID
		if rs.Scan(&c) == nil {
			forums[c.String()] = true
		}
	}
	rs.Close()
	out := map[string]bool{}
	for id, c := range all {
		if !c.Enabled {
			continue
		}
		if IsDashboard(id) {
			ok, err := dashboardRule(ctx, tx, p.Person, id, c)
			if err != nil {
				return nil, err
			}
			if ok {
				out[id] = true
			}
			continue
		}
		if c.See != "Specific groups" {
			out[id] = true
			continue
		}
		for _, g := range c.Groups {
			kind, ref, _ := strings.Cut(g, ":")
			if kind == "ugroup" && groups[ref] || kind == "forum" && forums[ref] || placeCol[kind] != "" && place[kind] == ref {
				out[id] = true
			}
		}
	}
	return out, nil
}

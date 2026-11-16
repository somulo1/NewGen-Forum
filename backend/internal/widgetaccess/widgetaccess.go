// Package widgetaccess decides, on the server, which widgets (and so which dashboards) a person may use. A widget the
// organization switched on is for everyone, or only for the groups chosen for it in Widget settings: people placed in a
// department, sector, branch or unit; members of a forum; or a user group (every role in it). The four dashboard widgets
// (departments, sectors, branches, units) are opened only by user groups. The sidebar's My department /
// My sector / My branch / My unit follow this, and the dashboards themselves refuse anyone else, even with a typed address.
package widgetaccess

import (
	"context"
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
		if c.See != "Specific groups" {
			out[id] = true
			continue
		}
		dash := placeCol[id] != "" // a dashboard widget: only its user groups open it (being placed somewhere does not)
		for _, g := range c.Groups {
			kind, ref, _ := strings.Cut(g, ":")
			switch {
			case dash:
				if kind == "ugroup" && groups[ref] {
					out[id] = true
				}
			case kind == "ugroup" && groups[ref], kind == "forum" && forums[ref], placeCol[kind] != "" && place[kind] == ref:
				out[id] = true
			}
		}
	}
	return out, nil
}

// May says whether the person may use one widget (a dashboard level: departments, sectors, branches, units).
func May(ctx context.Context, a *app.App, tx pgx.Tx, p *tenant.Principal, id string) (bool, error) {
	all, err := Allowed(ctx, a, tx, p)
	return all[id], err
}

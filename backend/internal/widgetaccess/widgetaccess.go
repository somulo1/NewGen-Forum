// Package widgetaccess decides, on the server, which widgets (and so which dashboards) a person may use. A widget is for a
// person when the organization switched it on and the person holds the widget's "see" right on Roles & rights (for a
// dashboard, "View dashboard"). A widget with no such right is for everyone once it is switched on.
//
// The sidebar (GET /me/widgets), the dashboards themselves (even at a typed address) and sending tickets on a place's behalf
// all ask this package, so they can never disagree.
package widgetaccess

import (
	"context"

	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/authz"
	"hub/internal/orgsettings"
	"hub/internal/tenant"
)

// Config is what Widget settings keeps for one widget.
type Config struct {
	Enabled  bool     `json:"enabled"`
	Title    string   `json:"title,omitempty"`
	Managers []string `json:"managers,omitempty"` // role keys that manage this widget's settings
	Items    int      `json:"items,omitempty"`    // how many entries it lists on Home (0: its own default)
}

var seeRight = map[string]string{
	"announcements": "View", "tickets": "View", "projects": "View", "files": "View", "awards": "View", "calendar": "View",
	"performance": "View", "memos": "Read", "newsletter": "Read", "jobposts": "Read", "sessions": "Join", "secureshare": "Create link",
	"departments": "View dashboard", "sectors": "View dashboard", "branches": "View dashboard", "units": "View dashboard",
}

// SeeRight names the right that shows a widget to a role ("" when the widget has none and is for everyone once switched on).
func SeeRight(id string) string { return seeRight[id] }

// Configs reads the organization's widget settings.
func Configs(ctx context.Context, a *app.App, tx pgx.Tx, p *tenant.Principal) (map[string]Config, error) {
	all := map[string]Config{}
	err := orgsettings.Get(ctx, a, tx, p.Org, "widgets", &all)
	return all, err
}

// Allowed lists the widget ids the person may use.
func Allowed(ctx context.Context, a *app.App, tx pgx.Tx, p *tenant.Principal) (map[string]bool, error) {
	all, err := Configs(ctx, a, tx, p)
	if err != nil {
		return nil, err
	}
	want := map[string][]string{}
	for id, c := range all {
		if r := seeRight[id]; c.Enabled && r != "" {
			want[id] = []string{r}
		}
	}
	held, err := authz.Effective(ctx, tx, p.Person, want)
	if err != nil {
		return nil, err
	}
	out := map[string]bool{}
	for id, c := range all {
		if c.Enabled && (seeRight[id] == "" || len(held[id]) > 0) {
			out[id] = true
		}
	}
	return out, nil
}

// Dashboard says whether the person may open the dashboards of one level (departments, sectors, branches, units).
func Dashboard(ctx context.Context, a *app.App, tx pgx.Tx, p *tenant.Principal, module string) (bool, error) {
	all, err := Allowed(ctx, a, tx, p)
	return all[module], err
}

package settings

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"

	"hub/internal/authz"
	"hub/internal/httpx"
	"hub/internal/orgsettings"
	"hub/internal/realtime"
	"hub/internal/tenant"
	"hub/internal/widgetaccess"
)

// Widget settings hold, per widget: whether the organization switched it on, its title, the roles that manage it and how many
// entries it lists on Home. Who sees a widget is not here: it is the widget's right on Roles & rights (see widgetaccess).
//
// "Manage widget settings" (Home & widgets) may change every widget, including who manages each one. A role named as a
// widget's manager may change that widget's title, switch and Home entries, and nothing else.

const maxHomeItems = 20

var (
	widgetID = regexp.MustCompile(`^[a-z][a-z0-9_]{0,31}$`)
	roleKey  = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)
)

// widgetPower says whether the person manages every widget, and otherwise which role they hold.
func widgetPower(ctx context.Context, tx pgx.Tx, p *tenant.Principal) (bool, string, error) {
	admin, err := authz.Can(ctx, tx, p.Person, "home", "Manage widget settings")
	if err != nil || admin {
		return admin, "", err
	}
	var role string
	err = tx.QueryRow(ctx, `select r.key from hub.person_roles pr join hub.roles r on r.id = pr.role_id where pr.person_id=$1`, p.Person).Scan(&role)
	if errors.Is(err, pgx.ErrNoRows) {
		err = nil
	}
	return false, role, err
}

// Widgets returns the configuration this person may read: all of it for a widget administrator; otherwise the switched-on
// widgets and the ones their role manages, so a widget never switched on does not reach anyone else's screen.
func (m *Module) Widgets(ctx context.Context, p *tenant.Principal) (map[string]widgetaccess.Config, error) {
	out := map[string]widgetaccess.Config{}
	err := m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		all, err := widgetaccess.Configs(ctx, m.A, tx, p)
		if err != nil {
			return err
		}
		admin, role, err := widgetPower(ctx, tx, p)
		if err != nil {
			return err
		}
		for id, c := range all {
			if admin || c.Enabled || (role != "" && slices.Contains(c.Managers, role)) {
				out[id] = c
			}
		}
		return nil
	})
	return out, err
}

// parseWidgets reads and checks a configuration sent by a page.
func parseWidgets(in map[string]any) (map[string]widgetaccess.Config, error) {
	if len(in) > 80 {
		return nil, httpx.Bad("That is too many widgets.")
	}
	out := map[string]widgetaccess.Config{}
	for id, v := range in {
		raw, _ := json.Marshal(v)
		var c widgetaccess.Config
		if _, isObj := v.(map[string]any); !isObj || !widgetID.MatchString(id) || json.Unmarshal(raw, &c) != nil {
			return nil, httpx.Bad("A widget setting is not valid.")
		}
		c.Title = strings.TrimSpace(c.Title)
		if utf8.RuneCountInString(c.Title) > 60 {
			return nil, httpx.Bad("Keep a widget's name to 60 characters.")
		}
		if c.Items < 0 || c.Items > maxHomeItems {
			return nil, httpx.Bad("Show between 1 and 20 entries on Home.")
		}
		if len(c.Managers) > 100 {
			return nil, httpx.Bad("That is too many roles.")
		}
		for _, k := range c.Managers {
			if !roleKey.MatchString(k) {
				return nil, httpx.Bad("Choose the roles that manage a widget from the list.")
			}
		}
		out[id] = c
	}
	return out, nil
}

func same(a, b widgetaccess.Config) bool {
	return a.Enabled == b.Enabled && a.Title == b.Title && a.Items == b.Items && slices.Equal(a.Managers, b.Managers)
}

// SetWidgets saves the widgets in the body; widgets left out keep their settings.
func (m *Module) SetWidgets(ctx context.Context, p *tenant.Principal, body map[string]any) error {
	in, err := parseWidgets(body)
	if err != nil {
		return err
	}
	var on, off []string
	err = m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		admin, role, err := widgetPower(ctx, tx, p)
		if err != nil {
			return err
		}
		all, err := widgetaccess.Configs(ctx, m.A, tx, p)
		if err != nil {
			return err
		}
		for id, c := range in {
			was := all[id]
			if same(was, c) {
				continue
			}
			mine := role != "" && slices.Contains(was.Managers, role)
			if !admin && (!mine || !slices.Equal(was.Managers, c.Managers)) {
				return httpx.ErrForbidden
			}
			if c.Enabled != was.Enabled {
				if c.Enabled {
					on = append(on, id)
				} else {
					off = append(off, id)
				}
			}
			all[id] = c
		}
		if err := orgsettings.Set(ctx, m.A, tx, p.Org, "widgets", all); err != nil {
			return err
		}
		slices.Sort(on)
		slices.Sort(off)
		return m.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Changed the home widget settings", "Settings", p.IP, map[string]any{"switched_on": on, "switched_off": off})
	})
	if err != nil {
		return err
	}
	// everyone signed in to this organization is told, so the change shows on their screen now, not at their next sign-in
	m.A.Publish(ctx, realtime.Event{Type: "widgets.changed", Channel: realtime.OrgChannel(p.Org), Org: p.Org.String(), Actor: p.Person.String()})
	return nil
}

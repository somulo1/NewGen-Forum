// Package security is the Security & logs page: the audit log, alerts raised by rules, and proof that the log was not altered.
package security

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/authz"
	"hub/internal/integrity"
	"hub/internal/modules/tickets"
	"hub/internal/notify"
)

type RuleDef struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Desc     string `json:"desc"`
	Severity string `json:"severity"`
	Default  bool   `json:"default_on"`
}

// Rules is the catalog of detections. An organization can switch each off or change its severity.
var Rules = []RuleDef{
	{"r1", "Repeated failed sign-ins", "5 or more failed sign-ins on one account within 10 minutes.", "High", true},
	{"r2", "Sign-in from a new country", "Someone signs in from a country they have never used before.", "High", true},
	{"r3", "Mass file download", "More than 50 files downloaded within 15 minutes.", "High", true},
	{"r4", "After-hours admin action", "Admin changes made outside working hours.", "Medium", true},
	{"r5", "Role or rights change", "Any change to roles, rights or approval rules.", "Medium", true},
	{"r6", "Secure link opened from an unusual IP", "A one-time link is opened from outside the usual IP ranges.", "Medium", true},
	{"r7", "Outside sharing spike", "More than 10 posts shared outside Hub within an hour.", "Low", false},
}

type ruleState struct {
	on  bool
	sev string
}

func loadRules(ctx context.Context, tx pgx.Tx) (map[string]ruleState, error) {
	m := map[string]ruleState{}
	for _, r := range Rules {
		m[r.ID] = ruleState{r.Default, r.Severity}
	}
	rs, err := tx.Query(ctx, `select id, enabled, severity from hub.security_rules`)
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	for rs.Next() {
		var id, sev string
		var on bool
		if err := rs.Scan(&id, &on, &sev); err != nil {
			return nil, err
		}
		m[id] = ruleState{on, sev}
	}
	return m, rs.Err()
}

var (
	outsideRe  = regexp.MustCompile(`^Shared (?:(\d+) (\w+) )?.*?outside Hub`)
	downloadRe = regexp.MustCompile(`^Downloaded (\d+) files`)
)

var atRe = regexp.MustCompile(` at \d\d:\d\d$`)

var rightsRe = regexp.MustCompile(`(?i)\b(rights?|roles?|approval)\b`)

// Evaluate is the audit hook: it looks at each new log line and raises an alert when it trips a rule. It runs in the same
// transaction as the action, so an alert and the thing that caused it are recorded together.
func Evaluate(ctx context.Context, a *app.App, tx pgx.Tx, e *app.Entry) error {
	rules, err := loadRules(ctx, tx)
	if err != nil {
		return err
	}
	var nb notify.Batch
	raise := func(rule, title, detail string) error {
		st := rules[rule]
		if !st.on {
			return nil
		}
		e.Rule = rule
		var seq int
		if err := tx.QueryRow(ctx, `insert into hub.counters (org_id, name, value) values ($1,'alert',112) on conflict (org_id, name) do update set value = hub.counters.value + 1 returning value`, e.Org).Scan(&seq); err != nil {
			return err
		}
		ref := fmt.Sprintf("AL-%d", seq)
		id := app.NewID()
		if _, err := tx.Exec(ctx, `insert into hub.alerts (org_id, id, ref, rule, severity, title, detail, ledger_seq) values ($1,$2,$3,$4,$5,$6,$7,$8)`, e.Org, id, ref, rule, st.sev, title, detail, e.Seq); err != nil {
			return err
		}
		if err := a.Sign.Seal(ctx, tx, "alerts", "id=$1", id); err != nil {
			return err
		}
		people, err := authz.PeopleWith(ctx, tx, "security", "View alerts")
		if err != nil {
			return err
		}
		for _, p := range people {
			if err := nb.Add(ctx, tx, e.Org, p, "alert", ref+": "+title, map[string]string{"High": "#FF9B8A", "Medium": "#FFB86B", "Low": "#7CC4FF"}[st.sev], ref); err != nil {
				return err
			}
		}
		return nil
	}
	defer func() { go nb.Send(context.Background(), a) }()

	switch {
	case e.Module == "Auth" && e.Result == "Failed":
		var n int
		if err := tx.QueryRow(ctx, `select count(*) from hub.audit_logs where actor_name=$1 and module='Auth' and result='Failed' and ts > coalesce($2, now()) - interval '10 minutes' and ts <= coalesce($2, now())`, e.Name, e.At).Scan(&n); err != nil {
			return err
		}
		if n+1 == 5 { // exactly the fifth: one alert per burst
			return raise("r1", fmt.Sprintf("5 failed sign-ins for %s within 10 minutes", e.Name), "From "+e.IP+". The account is locked after the fifth attempt.")
		}
	case e.Module == "Auth" && e.Result == "Success" && e.Location != "":
		cc := e.Location[strings.LastIndex(e.Location, ",")+1:]
		var seen bool
		if err := tx.QueryRow(ctx, `select exists (select 1 from hub.audit_logs where actor_name=$1 and module='Auth' and result='Success' and location like '%'||$2 and ts > now() - interval '180 days')`, e.Name, cc).Scan(&seen); err != nil {
			return err
		}
		var any bool
		if err := tx.QueryRow(ctx, `select exists (select 1 from hub.audit_logs where actor_name=$1 and module='Auth' and result='Success')`, e.Name).Scan(&any); err != nil {
			return err
		}
		if any && !seen {
			return raise("r2", e.Name+" signed in from "+e.Location+" for the first time", "IP "+e.IP+".")
		}
	case e.Module == "Roles" || e.Module == "People" || e.Module == "Settings":
		if e.Result == "Success" && e.Actor != nil {
			if e.Module == "Roles" && rightsRe.MatchString(e.Action) {
				if err := raise("r5", e.Action, "Made by "+e.Name+" from "+e.IP+". Logged automatically."); err != nil {
					return err
				}
			}
			if strings.HasPrefix(e.Action, "Created the organization") {
				return nil // setting an organization up is not a change to one that exists
			}
			cfg, err := tickets.LoadConfig(ctx, tx)
			if err != nil {
				return err
			}
			loc, _ := time.LoadLocation(cfg.Timezone)
			if loc == nil {
				loc = time.UTC
			}
			now := time.Now()
			if e.At != nil {
				now = *e.At
			}
			now = now.In(loc)
			m := now.Hour()*60 + now.Minute()
			from, to := minutes(cfg.Start, 8*60), minutes(cfg.End, 17*60)
			if m < from || m >= to {
				title := fmt.Sprintf("%s: %s at %s", e.Name, e.Action, now.Format("15:04"))
				if atRe.MatchString(e.Action) {
					title = e.Name + ": " + e.Action
				}
				return raise("r4", title, fmt.Sprintf("Admin action outside working hours (%s–%s).", cfg.Start, cfg.End))
			}
		}
	case e.Module == "Secure share" && strings.HasPrefix(e.Action, "Opened secure link"):
		if !usualIP(ctx, tx, e.IP) {
			return raise("r6", e.Action+" from "+e.IP, "The IP is outside the usual ranges.")
		}
	case outsideRe.MatchString(e.Action):
		// either one log line that says how many ("Shared 12 posts outside Hub in one hour") or single shares that add up
		if m := outsideRe.FindStringSubmatch(e.Action); m[1] != "" {
			if n, _ := strconv.Atoi(m[1]); n > 10 {
				return raise("r7", e.Name+" shared "+m[1]+" "+m[2]+" outside Hub in one hour", "The limit is 10 an hour.")
			}
			return nil
		}
		var n int
		if err := tx.QueryRow(ctx, `select count(*) from hub.audit_logs where actor_name=$1 and action like 'Shared%outside Hub%' and ts > coalesce($2, now()) - interval '1 hour' and ts <= coalesce($2, now())`, e.Name, e.At).Scan(&n); err != nil {
			return err
		}
		if n+1 == 11 {
			return raise("r7", e.Name+" shared 11 items outside Hub in one hour", "The limit is 10 an hour.")
		}
	case downloadRe.MatchString(e.Action):
		if n, _ := strconv.Atoi(downloadRe.FindStringSubmatch(e.Action)[1]); n > 50 {
			return raise("r3", e.Name+" downloaded "+strconv.Itoa(n)+" files in 15 minutes", "The limit is 50 files in 15 minutes.")
		}
	}
	return nil
}

func minutes(s string, def int) int {
	t, err := time.Parse("15:04", s)
	if err != nil {
		return def
	}
	return t.Hour()*60 + t.Minute()
}

// usualIP is true when the address shares its /16 with a successful sign-in in the last 30 days.
func usualIP(ctx context.Context, tx pgx.Tx, ip string) bool {
	if parsed := net.ParseIP(ip); parsed == nil || parsed.To4() == nil {
		return false
	}
	var seen bool
	_ = tx.QueryRow(ctx, `select exists (select 1 from hub.audit_logs where module='Auth' and result='Success' and ts > now() - interval '30 days'
		and case when ip ~ '^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$' then network(set_masklen(ip::inet, 16)) end = network(set_masklen($1::inet, 16)))`, ip).Scan(&seen)
	return seen
}

var _ = integrity.Bump
var _ = uuid.Nil

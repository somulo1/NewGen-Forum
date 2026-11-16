package security

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/authz"
	"hub/internal/crypto"
	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/tenant"
)

type Module struct{ A *app.App }

func New(a *app.App) *Module {
	m := &Module{A: a}
	a.Hooks = append(a.Hooks, Evaluate)
	return m
}

func (m *Module) Public(chi.Router) {}

type Alert struct {
	ID        uuid.UUID  `json:"id"`
	Ref       string     `json:"ref"`
	Rule      string     `json:"rule"`
	RuleName  string     `json:"rule_name"`
	Severity  string     `json:"severity"`
	Title     string     `json:"title"`
	Detail    string     `json:"detail"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	ClosedAt  *time.Time `json:"closed_at,omitempty"`
	Notes     []Note     `json:"notes"`
}

type Note struct {
	Status  string    `json:"status"`
	Comment string    `json:"comment"`
	Who     string    `json:"who"`
	At      time.Time `json:"at"`
}

func ruleName(id string) string {
	for _, r := range Rules {
		if r.ID == id {
			return r.Name
		}
	}
	return id
}

func (m *Module) Alerts(ctx context.Context, p *tenant.Principal, openOnly bool) ([]Alert, error) {
	out := []Alert{}
	err := m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "security", "View alerts"); err != nil {
			return err
		}
		dek, _, err := m.A.OrgDEK(ctx, tx, p.Org)
		if err != nil {
			return err
		}
		q := `select id, ref, rule, severity, title, detail, status, created_at, closed_at from hub.alerts`
		if openOnly {
			q += ` where status in ('Open','Investigating')`
		}
		rs, err := tx.Query(ctx, q+` order by created_at desc, id desc limit 300`)
		if err != nil {
			return err
		}
		for rs.Next() {
			var a Alert
			if err := rs.Scan(&a.ID, &a.Ref, &a.Rule, &a.Severity, &a.Title, &a.Detail, &a.Status, &a.CreatedAt, &a.ClosedAt); err != nil {
				rs.Close()
				return err
			}
			a.RuleName, a.Notes = ruleName(a.Rule), []Note{}
			out = append(out, a)
		}
		rs.Close()
		if len(out) > 0 {
			ids := make([]uuid.UUID, len(out))
			at := make(map[uuid.UUID]int, len(out))
			for i := range out {
				ids[i] = out[i].ID
				at[out[i].ID] = i
			}
			ns, err := tx.Query(ctx, `select n.alert_id, n.id, n.status, n.comment_enc, p.first_name || ' ' || p.last_name, n.at from hub.alert_notes n join hub.people p on p.id = n.person where n.alert_id = any($1) order by n.alert_id, n.at, n.id`, ids)
			if err != nil {
				return err
			}
			for ns.Next() {
				var aid, nid uuid.UUID
				var n Note
				var enc []byte
				if err := ns.Scan(&aid, &nid, &n.Status, &enc, &n.Who, &n.At); err != nil {
					ns.Close()
					return err
				}
				if b, err := crypto.FieldOpen(dek, noteCtx(p.Org, nid), enc); err == nil {
					n.Comment = string(b)
				}
				if i, ok := at[aid]; ok {
					out[i].Notes = append(out[i].Notes, n)
				}
			}
			ns.Close()
		}
		return nil
	})
	return out, err
}

func noteCtx(org, id uuid.UUID) crypto.Context {
	return crypto.Context{Org: org.String(), Table: "alert_notes", Column: "comment_enc", RowID: id.String()}
}

var nextStatus = map[string]bool{"Investigating": true, "Resolved": true, "False alarm": true}

// SetAlertStatus moves an alert along. A comment is required, except when closing as a false alarm.
func (m *Module) SetAlertStatus(ctx context.Context, p *tenant.Principal, id uuid.UUID, status, comment string) error {
	comment = strings.TrimSpace(comment)
	if !nextStatus[status] {
		return httpx.Bad("Choose Investigating, Resolved or False alarm.")
	}
	// a false alarm may be closed without a comment; everything else needs one
	if comment == "" && status == "False alarm" {
		comment = "Closed as a false alarm, no comment."
	}
	if len(comment) < 3 || len(comment) > 1000 {
		return httpx.Bad("Say what you found or did (3 to 1,000 characters).")
	}
	return m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "security", "Resolve alerts"); err != nil {
			return err
		}
		var cur, ref string
		if err := tx.QueryRow(ctx, `select status, ref from hub.alerts where id=$1 for update`, id).Scan(&cur, &ref); errors.Is(err, pgx.ErrNoRows) {
			return httpx.ErrNotFound
		} else if err != nil {
			return err
		}
		if cur == "Resolved" || cur == "False alarm" {
			return httpx.Err(409, "closed", "This alert is already closed.")
		}
		dek, _, err := m.A.OrgDEK(ctx, tx, p.Org)
		if err != nil {
			return err
		}
		nid := app.NewID()
		enc, _ := crypto.Field(dek, noteCtx(p.Org, nid), []byte(comment))
		if _, err := tx.Exec(ctx, `insert into hub.alert_notes (org_id, id, alert_id, status, comment_enc, person) values ($1,$2,$3,$4,$5,$6)`, p.Org, nid, id, status, enc, p.Person); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `update hub.alerts set status=$2, closed_at = case when $2 in ('Resolved','False alarm') then now() end, `+integrity.Bump+` where id=$1`, id, status); err != nil {
			return err
		}
		if err := m.A.Sign.Seal(ctx, tx, "alerts", "id=$1", id); err != nil {
			return err
		}
		return m.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, fmt.Sprintf("Marked %s as %s", ref, status), "Security", p.IP, map[string]any{"alert": id})
	})
}

type RuleOut struct {
	RuleDef
	Enabled bool   `json:"enabled"`
	Sev     string `json:"sev"`
}

func (m *Module) RulesFor(ctx context.Context, p *tenant.Principal) ([]RuleOut, error) {
	out := []RuleOut{}
	err := m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "security", "View alerts"); err != nil {
			return err
		}
		st, err := loadRules(ctx, tx)
		if err != nil {
			return err
		}
		for _, r := range Rules {
			out = append(out, RuleOut{r, st[r.ID].on, st[r.ID].sev})
		}
		return nil
	})
	return out, err
}

func (m *Module) SetRule(ctx context.Context, p *tenant.Principal, id string, enabled bool, sev string) error {
	known := false
	for _, r := range Rules {
		known = known || r.ID == id
	}
	if !known || (sev != "High" && sev != "Medium" && sev != "Low") {
		return httpx.Bad("Unknown rule or severity.")
	}
	return m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "security", "Manage rules"); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `insert into hub.security_rules (org_id, id, enabled, severity) values ($1,$2,$3,$4) on conflict (org_id, id) do update set enabled=excluded.enabled, severity=excluded.severity, `+integrity.BumpOn("security_rules"), p.Org, id, enabled, sev); err != nil {
			return err
		}
		if err := m.A.Sign.Seal(ctx, tx, "security_rules", "id=$1", id); err != nil {
			return err
		}
		return m.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, fmt.Sprintf("Set rule %s: %v, %s", id, enabled, sev), "Security", p.IP, nil)
	})
}

// ---- logs ----

type LogQuery struct {
	Q, Module, Result string
	Flagged           bool
	From, To          *time.Time
	Before            *time.Time
	Limit             int
}

type Log struct {
	ID       uuid.UUID `json:"id"`
	At       time.Time `json:"at"`
	User     string    `json:"user"`
	Action   string    `json:"action"`
	Module   string    `json:"module"`
	IP       string    `json:"ip"`
	Location string    `json:"location"`
	Result   string    `json:"result"`
	Rule     string    `json:"rule,omitempty"`
}

func (m *Module) Logs(ctx context.Context, p *tenant.Principal, q LogQuery, export bool) ([]Log, error) {
	if q.Limit <= 0 || q.Limit > 500 {
		q.Limit = 100
	}
	if export {
		q.Limit = 50000
	}
	out := []Log{}
	err := m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		right := "View logs"
		if export {
			right = "Export logs"
		}
		if err := authz.Require(ctx, tx, p.Person, "security", right); err != nil {
			return err
		}
		where, args := []string{"true"}, []any{}
		add := func(c string, v any) { args = append(args, v); where = append(where, fmt.Sprintf(c, len(args))) }
		if q.Module != "" {
			add("module = $%d", q.Module)
		}
		if q.Result != "" {
			add("result = $%d", q.Result)
		}
		if q.Flagged {
			where = append(where, "rule <> ''")
		}
		if q.Q != "" {
			args = append(args, "%"+strings.ToLower(escape(q.Q))+"%")
			where = append(where, fmt.Sprintf("(lower(actor_name) like $%d or lower(action) like $%d)", len(args), len(args)))
		}
		if q.From != nil {
			add("ts >= $%d", *q.From)
		}
		if q.To != nil {
			add("ts < $%d", *q.To)
		}
		if q.Before != nil {
			add("ts < $%d", *q.Before)
		}
		rs, err := tx.Query(ctx, `select id, ts, actor_name, action, module, ip, location, result, rule from hub.audit_logs where `+strings.Join(where, " and ")+fmt.Sprintf(" order by ts desc, id desc limit %d", q.Limit), args...)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var l Log
			if err := rs.Scan(&l.ID, &l.At, &l.User, &l.Action, &l.Module, &l.IP, &l.Location, &l.Result, &l.Rule); err != nil {
				return err
			}
			out = append(out, l)
		}
		return rs.Err()
	})
	return out, err
}

func escape(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

func parseQ(r *http.Request) (LogQuery, error) {
	v := r.URL.Query()
	q := LogQuery{Q: strings.TrimSpace(v.Get("q")), Module: v.Get("module"), Result: v.Get("result"), Flagged: v.Get("flagged") == "1"}
	q.Limit, _ = strconv.Atoi(v.Get("limit"))
	for k, dst := range map[string]**time.Time{"from": &q.From, "to": &q.To, "before": &q.Before} {
		if s := v.Get(k); s != "" {
			t, err := time.Parse(time.RFC3339, s)
			if err != nil {
				return q, httpx.Bad("Dates look like 2026-10-19T08:00:00Z.")
			}
			*dst = &t
		}
	}
	return q, nil
}

// Verify walks the organization's ledger and reports whether it is intact.
func (m *Module) Verify(ctx context.Context, p *tenant.Principal) (map[string]any, error) {
	out := map[string]any{"intact": true}
	err := m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "security", "View logs"); err != nil {
			return err
		}
		n, err := m.A.Ledger.Verify(ctx, tx, p.Org)
		out["entries"] = n
		var br *integrity.Break
		if errors.As(err, &br) {
			out["intact"], out["broken_at"], out["reason"] = false, br.Seq, br.Reason
			return nil
		}
		return err
	})
	return out, err
}

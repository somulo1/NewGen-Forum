package security

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/xuri/excelize/v2"

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

func (m *Module) Private(r chi.Router) {
	org := func(h func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error) http.HandlerFunc {
		return httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
			p, err := tenant.MustOrg(r.Context())
			if err != nil {
				return err
			}
			return h(w, r, p)
		})
	}
	r.Get("/security/activity", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		a, err := m.Activity(r.Context(), p)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, a)
		return nil
	}))
	r.Get("/security/alerts", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		l, err := m.Alerts(r.Context(), p, r.URL.Query().Get("open") == "1")
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"alerts": l})
		return nil
	}))
	r.Post("/security/alerts/{id}/status", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			return httpx.ErrNotFound
		}
		var in struct{ Status, Comment string }
		if err := httpx.Decode(r, &in, 4<<10); err != nil {
			return err
		}
		if err := m.SetAlertStatus(r.Context(), p, id, in.Status, in.Comment); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Get("/security/rules", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		l, err := m.RulesFor(r.Context(), p)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"rules": l})
		return nil
	}))
	r.Put("/security/rules/{id}", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		var in struct {
			Enabled  bool   `json:"enabled"`
			Severity string `json:"severity"`
		}
		if err := httpx.Decode(r, &in, 1<<10); err != nil {
			return err
		}
		if err := m.SetRule(r.Context(), p, chi.URLParam(r, "id"), in.Enabled, in.Severity); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Get("/security/blocks", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		days, _ := strconv.Atoi(r.URL.Query().Get("days"))
		l, err := m.Blocks(r.Context(), p, days)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"blocks": l})
		return nil
	}))
	r.Get("/security/logs", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		q, err := parseQ(r)
		if err != nil {
			return err
		}
		l, err := m.Logs(r.Context(), p, q, false)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"logs": l})
		return nil
	}))
	r.Get("/security/logs/{id}", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			return httpx.ErrNotFound
		}
		d, err := m.LogDetail(r.Context(), p, id)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, d)
		return nil
	}))
	r.Get("/security/logs/export", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		q, err := parseQ(r)
		if err != nil {
			return err
		}
		l, err := m.Logs(r.Context(), p, q, true)
		if err != nil {
			return err
		}
		// exporting is itself logged
		_ = m.A.DB.WithOrg(r.Context(), p.Org, func(tx pgx.Tx) error {
			return m.A.Audit(r.Context(), tx, p.Org, &p.Person, p.Name, fmt.Sprintf("Exported %d log lines", len(l)), "Security", p.IP, nil)
		})
		rows := [][]string{{"When", "Person", "What they did", "Where", "IP address", "Location", "Result", "Rule"}}
		for _, x := range l {
			rows = append(rows, []string{x.At.UTC().Format("2006-01-02 15:04:05"), x.User, x.Action, x.Module, x.IP, x.Location, x.Result, x.Rule})
		}
		if r.URL.Query().Get("format") == "csv" {
			var b bytes.Buffer
			cw := csv.NewWriter(&b)
			for _, row := range rows {
				for i, c := range row {
					row[i] = csvSafe(c)
				}
				_ = cw.Write(row)
			}
			cw.Flush()
			w.Header().Set("Content-Type", "text/csv; charset=utf-8")
			w.Header().Set("Content-Disposition", `attachment; filename="hub-logs.csv"`)
			_, _ = w.Write(b.Bytes())
			return nil
		}
		f := excelize.NewFile()
		defer f.Close()
		sheet := "Logs"
		_ = f.SetSheetName("Sheet1", sheet)
		for i, row := range rows {
			cells := make([]any, len(row))
			for j, c := range row {
				cells[j] = csvSafe(c)
			}
			cell, _ := excelize.CoordinatesToCellName(1, i+1)
			if err := f.SetSheetRow(sheet, cell, &cells); err != nil {
				return err
			}
		}
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", `attachment; filename="hub-logs.xlsx"`)
		return f.Write(w)
	}))
	r.Get("/security/integrity", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		out, err := m.Verify(r.Context(), p)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, out)
		return nil
	}))
}

// csvSafe stops a spreadsheet from running a cell that starts with = + - @ as a formula (CSV/Excel injection).
func csvSafe(s string) string {
	if s != "" && strings.ContainsRune("=+-@\t\r", rune(s[0])) {
		return "'" + s
	}
	return s
}

// Block is one time sign-in was stopped. Blocks lift by themselves; this is the record.
type Block struct {
	At     time.Time `json:"at"`
	Until  time.Time `json:"until"`
	Kind   string    `json:"kind"`   // account | address
	Member bool      `json:"member"` // the person belongs to this organization
	Who    string    `json:"who"`    // their name, or a masked address for someone who is not a member
	IP     string    `json:"ip"`
	Reason string    `json:"reason"`
}

// Blocks lists the sign-in blocks of the last days (default 7): people of this organization, and also addresses blocked for
// trying passwords on accounts that are not members (their email is masked).
func (m *Module) Blocks(ctx context.Context, p *tenant.Principal, days int) ([]Block, error) {
	if days <= 0 || days > 90 {
		days = 7
	}
	err := m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		return authz.Require(ctx, tx, p.Person, "security", "View logs")
	})
	if err != nil {
		return nil, err
	}
	out := []Block{}
	err = m.A.DB.Tx(ctx, func(tx pgx.Tx) error {
		rs, err := tx.Query(ctx, `select b.at, b.until, b.kind, b.ip, b.reason, b.email_hint,
				exists (select 1 from hub.account_orgs ao where ao.account_id = b.account_id and ao.org_id = $2), b.account_id is not null, coalesce(a.display_name, '')
			from hub.login_blocks b left join hub.accounts a on a.id = b.account_id
			where b.at > now() - make_interval(days => $1)
				and (b.account_id is null or exists (select 1 from hub.account_orgs ao where ao.account_id = b.account_id and ao.org_id = $2))
			order by b.at desc limit 1000`, days, p.Org)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var b Block
			var hint, name string
			var hasAcct bool
			if err := rs.Scan(&b.At, &b.Until, &b.Kind, &b.IP, &b.Reason, &hint, &b.Member, &hasAcct, &name); err != nil {
				return err
			}
			b.Who = hint
			if b.Member && name != "" {
				b.Who = name
			}
			out = append(out, b)
		}
		return rs.Err()
	})
	return out, err
}

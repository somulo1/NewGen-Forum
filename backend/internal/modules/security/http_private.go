package security

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/xuri/excelize/v2"

	"hub/internal/authz"
	"hub/internal/httpx"
	"hub/internal/tenant"
)

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

package security

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/integrity"
)

// ScanIntegrity re-checks every sealed table and ledger (with the read-only audit role) and turns anything changed
// outside the system into an audit entry per organization, which raises the "Tampered records" alert (r8) on the security
// logs page. A problem already reported in the last day is not reported again.
func ScanIntegrity(ctx context.Context, a *app.App, auditURL string) (int, error) {
	conn, err := pgx.Connect(ctx, auditURL)
	if err != nil {
		return 0, fmt.Errorf("integrity scan: cannot connect with the audit role: %w", err)
	}
	defer conn.Close(ctx)
	sc := &integrity.Scanner{S: a.Sign, L: a.Ledger, MaxBad: 1000}
	rep, err := sc.Run(ctx, conn, "")
	if err != nil {
		return 0, err
	}
	byOrg := map[string][]integrity.Problem{}
	schema := map[string][]integrity.Problem{}
	for _, p := range rep.Problems {
		h := sha256.Sum256([]byte(p.Table + "|" + p.Org + "|" + p.Row + "|" + p.Reason))
		if p.Cause == "schema" {
			schema[p.Org] = append(schema[p.Org], p) // authentic records: they are re-sealed below, not reported as tampering
			continue
		}
		if ok, err := a.Rdb.SetNX(ctx, "tamper-scan:"+hex.EncodeToString(h[:16]), 1, 24*time.Hour).Result(); err == nil && !ok {
			continue
		}
		byOrg[p.Org] = append(byOrg[p.Org], p)
	}
	// A database update that added a column makes older records look different without changing them. They verify without the
	// new column, so they are authentic: seal them again under the current form and say so.
	for org, list := range schema {
		id, err := uuid.Parse(org)
		if err != nil {
			continue
		}
		done := map[string]int{}
		cols := map[string]map[string]bool{}
		_ = a.DB.WithOrg(ctx, id, func(tx pgx.Tx) error {
			for _, p := range list {
				if resealOne(ctx, a, tx, p) == nil {
					done[p.Table]++
					if cols[p.Table] == nil {
						cols[p.Table] = map[string]bool{}
					}
					for _, c := range p.Added {
						cols[p.Table][c] = true
					}
				}
			}
			return nil
		})
		if len(done) == 0 {
			continue
		}
		var parts []string
		for t, n := range done {
			var cs []string
			for c := range cols[t] {
				cs = append(cs, c)
			}
			sort.Strings(cs)
			parts = append(parts, fmt.Sprintf("%s (%d, new column %s)", tableName(t), n, strings.Join(cs, ", ")))
		}
		sort.Strings(parts)
		_ = a.DB.WithOrg(ctx, id, func(tx pgx.Tx) error {
			return a.AuditEntry(ctx, tx, &app.Entry{Org: id, Name: "Integrity monitor", Module: "Integrity", Result: "Success",
				Action: "Records sealed again after a database update added a column: " + strings.Join(parts, "; ") + ". Nothing was changed outside Hub."})
		})
	}
	reported := 0
	for org, list := range byOrg {
		id, err := uuid.Parse(org)
		if err != nil {
			continue // a table shared by all organizations: the CLI report (cmd/integrity) covers it
		}
		detail := Describe(list)
		if err := a.DB.WithOrg(ctx, id, func(tx pgx.Tx) error {
			return a.AuditEntry(ctx, tx, &app.Entry{Org: id, Name: "Integrity monitor", Module: "Integrity", Result: "Failed",
				Action:  Headline(list),
				Payload: map[string]any{"detail": detail, "problems": list}})
		}); err != nil {
			return reported, err
		}
		reported += len(list)
	}
	return reported, nil
}

// resealOne gives an authentic record a new seal under the table's current form.
func resealOne(ctx context.Context, a *app.App, tx pgx.Tx, p integrity.Problem) error {
	if len(p.PK) == 0 || !identRe.MatchString(p.Table) {
		return fmt.Errorf("no key")
	}
	var where []string
	var args []any
	for c, v := range p.PK {
		if !identRe.MatchString(c) {
			return fmt.Errorf("bad column")
		}
		args = append(args, v)
		where = append(where, fmt.Sprintf("%s::text = $%d", c, len(args)))
	}
	w := strings.Join(where, " and ")
	if _, err := tx.Exec(ctx, `update hub.`+p.Table+` set `+integrity.Bump+` where `+w, args...); err != nil {
		return err
	}
	return a.Sign.Seal(ctx, tx, p.Table, w, args...)
}

var identRe = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)

var tableNames = map[string]string{"comms": "communications", "messages": "messages", "people": "people", "roles": "roles", "role_rights": "role rights", "person_rights": "personal rights",
	"person_roles": "role assignments", "tickets": "tickets", "files": "files", "records": "notes, tasks and calendar entries", "conversations": "conversations", "alerts": "security alerts",
	"approval_requests": "approval requests", "org_units": "organization units", "projects": "projects", "shares": "secure links", "orgs": "organizations", "accounts": "accounts"}

func tableName(t string) string {
	if n, ok := tableNames[t]; ok {
		return n
	}
	return t
}

// Headline is the alert title: how many records, in what.
func Headline(list []integrity.Problem) string {
	by := map[string]int{}
	for _, p := range list {
		by[p.Table]++
	}
	var parts []string
	for t, n := range by {
		parts = append(parts, fmt.Sprintf("%d %s", n, tableName(t)))
	}
	sort.Strings(parts)
	return fmt.Sprintf("%d record(s) were changed outside Hub: %s", len(list), strings.Join(parts, ", "))
}

// Describe says exactly which records no longer match their seal, one line each, and what to do about it.
func Describe(list []integrity.Problem) string {
	var b strings.Builder
	shown := 0
	for _, p := range list {
		if shown == 12 {
			fmt.Fprintf(&b, "…and %d more.\n", len(list)-shown)
			break
		}
		shown++
		what := "changed after it was written"
		if p.Cause == "unsealed" {
			what = "never got its seal"
		}
		fmt.Fprintf(&b, "• %s · %s: %s\n", tableName(p.Table), identLine(p), what)
	}
	b.WriteString("\nThese records are hidden from everyone until they are put right. Restore them from a backup and find out who had direct access to the database.")
	return b.String()
}

func identLine(p integrity.Problem) string {
	var parts []string
	for _, k := range []string{"ref", "name", "key", "type", "content_type", "kind", "collection", "status", "module", "role", "created_at", "ts"} {
		if v, ok := p.Ident[k]; ok && v != "" {
			if k == "created_at" || k == "ts" {
				if t, err := time.Parse(time.RFC3339Nano, v); err == nil {
					v = "from " + t.In(time.FixedZone("EAT", 3*3600)).Format("2 Jan 15:04")
				}
			}
			parts = append(parts, v)
		}
	}
	id := p.Ident["id"]
	if id == "" {
		for _, v := range p.PK {
			id = v
			break
		}
	}
	out := strings.Join(parts, ", ")
	if id != "" {
		out += " [id " + id + "]"
	}
	return strings.TrimSpace(out)
}

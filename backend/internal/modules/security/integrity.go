package security

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
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
	for _, p := range rep.Problems {
		h := sha256.Sum256([]byte(p.Table + "|" + p.Org + "|" + p.Row + "|" + p.Reason))
		if ok, err := a.Rdb.SetNX(ctx, "tamper-scan:"+hex.EncodeToString(h[:16]), 1, 24*time.Hour).Result(); err == nil && !ok {
			continue
		}
		byOrg[p.Org] = append(byOrg[p.Org], p)
	}
	reported := 0
	for org, list := range byOrg {
		id, err := uuid.Parse(org)
		if err != nil {
			continue // a table shared by all organizations: the CLI report (cmd/integrity) covers it
		}
		tables := map[string]int{}
		for _, p := range list {
			tables[p.Table]++
		}
		names := make([]string, 0, len(tables))
		for t, n := range tables {
			names = append(names, fmt.Sprintf("%s (%d)", t, n))
		}
		sort.Strings(names)
		if err := a.DB.WithOrg(ctx, id, func(tx pgx.Tx) error {
			return a.AuditEntry(ctx, tx, &app.Entry{Org: id, Name: "Integrity monitor", Module: "Integrity", Result: "Failed",
				Action:  fmt.Sprintf("%d record(s) were changed outside Hub: %s", len(list), strings.Join(names, ", ")),
				Payload: list})
		}); err != nil {
			return reported, err
		}
		reported += len(list)
	}
	return reported, nil
}

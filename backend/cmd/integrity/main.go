// Command integrity re-checks the whole database against the MAC key and the ledger chains, and exits non-zero if anything
// was changed outside the system. Run it on a schedule with the read-only audit role.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"

	"hub/internal/config"
	"hub/internal/integrity"
)

func main() {
	only := flag.String("table", "", "scan one table only")
	asJSON := flag.Bool("json", false, "print the report as JSON")
	flag.Parse()
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "integrity:", err)
		os.Exit(1)
	}
	url := os.Getenv("HUB_AUDIT_URL")
	if url == "" {
		url = "postgres://hub_audit:hub_audit_dev_pw@127.0.0.1:5433/hub?sslmode=disable"
	}
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		fmt.Fprintln(os.Stderr, "integrity: cannot connect with the audit role:", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)
	sc := &integrity.Scanner{S: integrity.NewSigner(cfg.MACKey), L: integrity.NewLedger(cfg.LedgerKey)}
	rep, err := sc.Run(ctx, conn, *only)
	if err != nil {
		fmt.Fprintln(os.Stderr, "integrity:", err)
		os.Exit(1)
	}
	if *asJSON {
		_ = json.NewEncoder(os.Stdout).Encode(rep)
	} else {
		fmt.Printf("checked %d tables, %d rows, %d organizations, %d ledgers\n", rep.Tables, rep.Rows, rep.Orgs, rep.Ledgers)
		for _, p := range rep.Problems {
			fmt.Printf("PROBLEM %s org=%s: %s\n  %s\n", p.Table, p.Org, p.Reason, p.Row)
		}
		if rep.OK() {
			fmt.Println("everything matches")
		}
	}
	if !rep.OK() {
		os.Exit(2)
	}
}

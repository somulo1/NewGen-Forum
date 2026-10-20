package main

import (
	"bytes"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/access"
	"hub/internal/app"
	"hub/internal/config"
)

// protect checks, without changing anything except an organization's missing data key, that the protections the design relies
// on are really in place in this database and configuration. Run it after every deployment and migration.
func protect(ctx context.Context, a *app.App, cfg *config.Config, slug string) int {
	bad := 0
	ok := func(good bool, what string) {
		mark := "ok  "
		if !good {
			mark, bad = "FAIL", bad+1
		}
		fmt.Printf("%s %s\n", mark, what)
	}
	// keys: set, and all different from each other (in prod the config already refuses missing ones)
	keys := map[string][]byte{"HUB_KEK": cfg.KEK, "HUB_MAC_KEY": cfg.MACKey, "HUB_BLIND_KEY": cfg.BlindKey, "HUB_TOKEN_KEY": cfg.TokenKey, "HUB_LEDGER_KEY": cfg.LedgerKey, "HUB_CALLS_SECRET": cfg.CallsSecret}
	distinct := true
	for n1, k1 := range keys {
		for n2, k2 := range keys {
			if n1 < n2 && bytes.Equal(k1, k2) {
				distinct = false
			}
		}
	}
	ok(distinct, "the six encryption, signing and call keys are all different")
	if cfg.Env == "prod" {
		ok(cfg.DBGuard != "dev-db-guard-token-change-me", "the database guard token is not the development default")
	}
	err := a.DB.Tx(ctx, func(tx pgx.Tx) error {
		var unprotected int
		_ = tx.QueryRow(ctx, `select count(*) from pg_class c join pg_namespace n on n.oid=c.relnamespace
			where n.nspname='hub' and c.relkind in ('r','p') and not c.relispartition
			and exists (select 1 from pg_attribute at where at.attrelid=c.oid and at.attname='org_id' and not at.attisdropped)
			and not (c.relrowsecurity and (c.relforcerowsecurity or c.relname in ('guard','migrations')))`).Scan(&unprotected)
		ok(unprotected == 0, fmt.Sprintf("row-level security is on for every table in the hub schema (%d without)", unprotected))
		var open int
		_ = tx.QueryRow(ctx, `select count(*) from pg_class c join pg_namespace n on n.oid=c.relnamespace where n.nspname='hub' and c.relkind in ('r','p') and not c.relispartition and c.relname not in ('guard','migrations')
			and not exists (select 1 from pg_policy p where p.polrelid=c.oid and (pg_get_expr(p.polqual, p.polrelid) like '%current_org()%' or pg_get_expr(p.polqual, p.polrelid) like '%is_app()%'))`).Scan(&open)
		ok(open == 0, fmt.Sprintf("every table is filtered by organization or closed to connections that are not the Hub system (%d not)", open))
		var direct int
		_ = tx.QueryRow(ctx, `select count(*) from pg_class c join pg_namespace n on n.oid=c.relnamespace where n.nspname='hub' and c.relispartition and has_table_privilege('hub_app', c.oid, 'select')`).Scan(&direct)
		ok(direct == 0, fmt.Sprintf("the application role has no direct access to table partitions (%d with)", direct))
		var super, bypass bool
		_ = tx.QueryRow(ctx, `select rolsuper, rolbypassrls from pg_roles where rolname = current_user`).Scan(&super, &bypass)
		ok(!super && !bypass, "the application connects as a role that is not a superuser and cannot bypass row-level security")
		var mut int
		_ = tx.QueryRow(ctx, `select count(*) from (values ('hub.ledger'),('hub.audit_logs'),('hub.org_keys')) t(n) where has_table_privilege(current_user, n, 'update') or has_table_privilege(current_user, n, 'delete')`).Scan(&mut)
		ok(mut == 0, "the audit ledger, audit log and data keys are append-only for the application role")
		return nil
	})
	if err != nil {
		fmt.Println("FAIL could not read the database:", err)
		return 1
	}
	// every organization has the full catalog of pages and rights, so any right can be given to any role
	var orgIDs []uuid.UUID
	_ = a.DB.Tx(ctx, func(tx pgx.Tx) error {
		rs, err := tx.Query(ctx, `select id from hub.orgs where status='active'`)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var id uuid.UUID
			if rs.Scan(&id) == nil {
				orgIDs = append(orgIDs, id)
			}
		}
		return nil
	})
	incomplete := 0
	for _, id := range orgIDs {
		_ = a.DB.WithOrg(ctx, id, func(tx pgx.Tx) error {
			if have, err := access.Installed(ctx, tx); err != nil || !have {
				incomplete++
			}
			return nil
		})
	}
	ok(incomplete == 0, fmt.Sprintf("all %d active organizations have the complete catalog of pages and rights (%d incomplete: run `bootstrap org -all`)", len(orgIDs), incomplete))
	if slug != "" {
		var org uuid.UUID
		if err := a.DB.Tx(ctx, func(tx pgx.Tx) error {
			return tx.QueryRow(ctx, `select id from hub.orgs where slug=$1`, slug).Scan(&org)
		}); err != nil {
			fmt.Printf("FAIL no organization with web name %q\n", slug)
			return 1
		}
		err := a.DB.WithOrg(ctx, org, func(tx pgx.Tx) error {
			dek, ver, err := a.OrgDEK(ctx, tx, org) // creates the organization's data key when it has none
			if err != nil {
				return err
			}
			ok(len(dek) == 32 && ver >= 1, fmt.Sprintf("organization %s has its own data key (version %d)", slug, ver))
			return nil
		})
		if err != nil {
			fmt.Println("FAIL organization data key:", err)
			bad++
		}
	}
	if bad > 0 {
		fmt.Printf("\n%d protection(s) missing.\n", bad)
		return 1
	}
	fmt.Println("\nall protections in place.")
	return 0
}

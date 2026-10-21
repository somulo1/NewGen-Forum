package db

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// Migrate applies the embedded SQL files with the owner role. Every applied file's SHA-256 is recorded; if a file
// that was already applied has changed, Migrate stops: schema history must never be rewritten.
func Migrate(ctx context.Context, ownerURL, guardToken string, files fs.FS) ([]string, error) {
	cctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	conn, err := pgx.Connect(cctx, ownerURL)
	if err != nil {
		return nil, fmt.Errorf("owner connection (HUB_MIGRATE_URL): %w", err)
	}
	defer conn.Close(ctx)
	// several nodes starting together must not run the same migration twice: one waits for the other
	if _, err := conn.Exec(ctx, `select pg_advisory_lock(727274)`); err != nil {
		return nil, err
	}
	defer func() { _, _ = conn.Exec(context.Background(), `select pg_advisory_unlock(727274)`) }()
	if _, err := conn.Exec(ctx, `create schema if not exists hub;
create table if not exists hub.migrations (version int primary key, name text not null, sha256 text not null, applied_at timestamptz not null default now())`); err != nil {
		return nil, err
	}
	names, err := fs.Glob(files, "*.sql")
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	var applied []string
	for _, name := range names {
		var ver int
		if _, err := fmt.Sscanf(name, "%d_", &ver); err != nil {
			return nil, fmt.Errorf("migration %s: name must start with a number", name)
		}
		body, err := fs.ReadFile(files, name)
		if err != nil {
			return nil, err
		}
		sum := sha256.Sum256(body)
		h := hex.EncodeToString(sum[:])
		var old string
		err = conn.QueryRow(ctx, "select sha256 from hub.migrations where version=$1", ver).Scan(&old)
		if err == nil {
			if old != h {
				return applied, fmt.Errorf("migration %s was changed after it was applied (expected %s, found %s): add a new migration instead", name, old[:12], h[:12])
			}
			continue
		}
		if err != pgx.ErrNoRows {
			return applied, err
		}
		tx, err := conn.Begin(ctx)
		if err != nil {
			return applied, err
		}
		if _, err := tx.Exec(ctx, string(body)); err != nil {
			_ = tx.Rollback(ctx)
			return applied, fmt.Errorf("migration %s: %w", name, err)
		}
		if _, err := tx.Exec(ctx, "insert into hub.migrations (version, name, sha256) values ($1,$2,$3)", ver, strings.TrimSuffix(name, ".sql"), h); err != nil {
			_ = tx.Rollback(ctx)
			return applied, err
		}
		if err := tx.Commit(ctx); err != nil {
			return applied, err
		}
		applied = append(applied, name)
	}
	return applied, applyGrants(ctx, conn, guardToken)
}

// applyGrants is idempotent: it sets who can do what after every migration run.
func applyGrants(ctx context.Context, conn *pgx.Conn, guardToken string) error {
	stmts := []string{
		`grant usage on schema hub to hub_app, hub_audit, hub_dba`,
		`grant select, insert, update, delete on all tables in schema hub to hub_app`,
		// append-only tables: the runtime role may only add rows
		`revoke update, delete, truncate on hub.ledger, hub.ledger_checkpoints, hub.audit_logs, hub.org_keys from hub_app`,
		`revoke all on hub.guard, hub.migrations from hub_app, hub_audit, hub_dba`,
		`grant select on all tables in schema hub to hub_audit`,
		`revoke select on hub.guard, hub.migrations from hub_audit`,
		`grant execute on all functions in schema hub to hub_app, hub_audit`,
		// humans (hub_dba) can look at structure through the catalog but get no table access at all
		`revoke all on all tables in schema hub from hub_dba`,
		`alter default privileges for role hub_owner in schema hub grant select, insert, update, delete on tables to hub_app`,
		`alter default privileges for role hub_owner in schema hub grant select on tables to hub_audit`,
		// Partitions (messages_p0.., records_p0..) get no direct access: row-level security lives on the parent and PostgreSQL
		// does not apply it to a query that names a partition. Access through the parent is unaffected.
		`do $$ declare t record; begin for t in select n.nspname, c.relname from pg_class c join pg_namespace n on n.oid = c.relnamespace where n.nspname = 'hub' and c.relispartition and c.relkind = 'r' loop execute format('revoke all on %I.%I from hub_app, hub_audit, hub_dba', t.nspname, t.relname); end loop; end $$`,
	}
	for _, s := range stmts {
		if _, err := conn.Exec(ctx, s); err != nil {
			return fmt.Errorf("grants: %s: %w", s, err)
		}
	}
	// The guard token is stored only as a hash; the plaintext lives in the service's secret store.
	_, err := conn.Exec(ctx, `insert into hub.guard (id, token_hash) values (1, public.digest($1, 'sha256')) on conflict (id) do update set token_hash = excluded.token_hash`, guardToken)
	return err
}

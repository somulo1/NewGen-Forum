package settings

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/authz"
	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/tenant"
)

// Deleting an organization. The request starts a 30-day wait (it can be cancelled in that time); then the worker erases
// everything: every record of the organization, its keys (so any copy of its encrypted data, such as a backup, can no
// longer be opened) and its files on every file server.

const DeletionWait = 30 * 24 * time.Hour

// RequestDeletion schedules the organization for deletion. confirm must be the organization's exact name.
func (m *Module) RequestDeletion(ctx context.Context, p *tenant.Principal, confirm string) (time.Time, error) {
	var due time.Time
	err := m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "settings", "Manage subscription"); err != nil {
			return err
		}
		var name string
		var cur *time.Time
		if err := tx.QueryRow(ctx, `select name, delete_after from hub.orgs where id=$1 for update`, p.Org).Scan(&name, &cur); err != nil {
			return err
		}
		if strings.TrimSpace(confirm) != name {
			return httpx.Bad("Type the organization's name exactly to confirm.")
		}
		if cur != nil {
			due = *cur
			return nil
		}
		due = time.Now().Add(DeletionWait).UTC()
		if _, err := tx.Exec(ctx, `update hub.orgs set delete_after=$2, updated_at=now(), `+integrity.Bump+` where id=$1`, p.Org, due); err != nil {
			return err
		}
		if err := m.A.Sign.Seal(ctx, tx, "orgs", "id=$1", p.Org); err != nil {
			return err
		}
		return m.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Requested deletion of the organization and all its data on "+due.Format("2 Jan 2006"), "Settings", p.IP, map[string]any{"delete_after": due})
	})
	return due, err
}

// CancelDeletion stops a scheduled deletion.
func (m *Module) CancelDeletion(ctx context.Context, p *tenant.Principal) error {
	return m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "settings", "Manage subscription"); err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, `update hub.orgs set delete_after=null, updated_at=now(), `+integrity.Bump+` where id=$1 and delete_after is not null`, p.Org)
		if err != nil || tag.RowsAffected() == 0 {
			return err
		}
		if err := m.A.Sign.Seal(ctx, tx, "orgs", "id=$1", p.Org); err != nil {
			return err
		}
		return m.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Cancelled the deletion of the organization", "Settings", p.IP, nil)
	})
}

// DeletionDue lists organizations whose deletion date has passed.
func (m *Module) DeletionDue(ctx context.Context) ([]uuid.UUID, error) {
	var out []uuid.UUID
	err := m.A.DB.Tx(ctx, func(tx pgx.Tx) error {
		rs, err := tx.Query(ctx, `select id from hub.orgs where delete_after <= now() and status <> 'deleted'`)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var id uuid.UUID
			if err := rs.Scan(&id); err != nil {
				return err
			}
			out = append(out, id)
		}
		return rs.Err()
	})
	return out, err
}

// FileRefs lists the stored contents of an organization's files (to erase them before its records go).
func (m *Module) FileRefs(ctx context.Context, org uuid.UUID) ([]string, error) {
	var out []string
	err := m.A.DB.WithOrg(ctx, org, func(tx pgx.Tx) error {
		rs, err := tx.Query(ctx, `select blob_key from hub.files where blob_key is not null union select ref from hub.blob_purge where org_id=$1`, org)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var k string
			if err := rs.Scan(&k); err != nil {
				return err
			}
			out = append(out, k)
		}
		return rs.Err()
	})
	return out, err
}

// EraseOrg removes every record of the organization, including its keys and logs, and marks it deleted. Call it after
// its files are erased. Tables are emptied in rounds so references between them resolve in any order.
func (m *Module) EraseOrg(ctx context.Context, org uuid.UUID) error {
	return m.A.DB.Tx(ctx, func(tx pgx.Tx) error {
		var due *time.Time
		if err := tx.QueryRow(ctx, `select delete_after from hub.orgs where id=$1 for update`, org).Scan(&due); err != nil {
			return err
		}
		if due == nil || due.After(time.Now()) {
			return fmt.Errorf("organization %s is not due for deletion", org)
		}
		for _, q := range []string{`select set_config('hub.org_id', $1, true)`, `select set_config('hub.purge_org', $1, true)`} {
			if _, err := tx.Exec(ctx, q, org.String()); err != nil {
				return err
			}
		}
		// logs, ledger and key history can only be erased through the database function that checks the date
		if _, err := tx.Exec(ctx, `select hub.erase_append_only($1)`, org); err != nil {
			return err
		}
		rs, err := tx.Query(ctx, `select c.relname from pg_class c join pg_namespace n on n.oid=c.relnamespace
			where n.nspname='hub' and c.relkind in ('r','p') and not c.relispartition and c.relname <> 'orgs'
			  and c.relname not in ('ledger','ledger_checkpoints','audit_logs','org_keys')
			  and exists (select 1 from pg_attribute a where a.attrelid=c.oid and a.attname='org_id' and not a.attisdropped)`)
		if err != nil {
			return err
		}
		var tables []string
		for rs.Next() {
			var t string
			if err := rs.Scan(&t); err != nil {
				rs.Close()
				return err
			}
			tables = append(tables, t)
		}
		rs.Close()
		left := tables
		lastErr := map[string]error{}
		for round := 0; round < 20 && len(left) > 0; round++ {
			var again []string
			for _, t := range left {
				if _, err := tx.Exec(ctx, `savepoint erase`); err != nil {
					return err
				}
				if _, err := tx.Exec(ctx, `delete from hub.`+t+` where org_id=$1`, org); err != nil {
					if _, err := tx.Exec(ctx, `rollback to savepoint erase`); err != nil {
						return err
					}
					lastErr[t] = err
					again = append(again, t)
					continue
				}
				if _, err := tx.Exec(ctx, `release savepoint erase`); err != nil {
					return err
				}
			}
			left = again
		}
		if len(left) > 0 {
			var why []string
			for _, t := range left {
				why = append(why, fmt.Sprintf("%s (%v)", t, lastErr[t]))
			}
			return fmt.Errorf("could not erase %s", strings.Join(why, "; "))
		}
		if _, err := tx.Exec(ctx, `update hub.orgs set status='deleted', name='Deleted organization', updated_at=now(), `+integrity.Bump+` where id=$1`, org); err != nil {
			return err
		}
		if err := m.A.Sign.Seal(ctx, tx, "orgs", "id=$1", org); err != nil {
			return err
		}
		slog.Info("organization erased", "org", org)
		return nil
	})
}

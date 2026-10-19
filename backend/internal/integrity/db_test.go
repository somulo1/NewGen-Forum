package integrity_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/config"
	"hub/internal/db"
	"hub/internal/integrity"
)

func setup(t *testing.T) (*config.Config, *db.DB, *integrity.Signer, *integrity.Ledger) {
	t.Helper()
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	d, err := db.Open(context.Background(), cfg.DatabaseURL, cfg.DBGuard, 4)
	if err != nil {
		t.Skip("database not available: ", err)
	}
	t.Cleanup(d.Close)
	return cfg, d, integrity.NewSigner(cfg.MACKey), integrity.NewLedger(cfg.LedgerKey)
}

func newOrg(t *testing.T, d *db.DB, s *integrity.Signer) uuid.UUID {
	t.Helper()
	id := uuid.Must(uuid.NewV7())
	slug := "t-" + strings.ToLower(id.String()[24:])
	err := d.Tx(context.Background(), func(tx pgx.Tx) error {
		if _, err := tx.Exec(context.Background(), `insert into hub.orgs (id, slug, name) values ($1,$2,$3)`, id, slug, "Test "+slug); err != nil {
			return err
		}
		return s.Seal(context.Background(), tx, "orgs", "id=$1", id)
	})
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func TestWritesWithoutTheGuardTokenAreRefused(t *testing.T) {
	cfg, _, _, _ := setup(t)
	d, err := db.Open(context.Background(), cfg.DatabaseURL, "wrong-token", 2)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	id := uuid.Must(uuid.NewV7())
	err = d.Tx(context.Background(), func(tx pgx.Tx) error {
		_, err := tx.Exec(context.Background(), `insert into hub.orgs (id, slug, name) values ($1,'sneaky','x')`, id)
		return err
	})
	if err == nil || !strings.Contains(err.Error(), "may only be written by the Hub system") {
		t.Fatalf("a connection without the guard token wrote data: %v", err)
	}
}

func TestRowNeedsMACAndVersionBump(t *testing.T) {
	_, d, s, _ := setup(t)
	ctx := context.Background()
	org := newOrg(t, d, s)
	// a forged MAC supplied by a script that does not know the key is accepted by the database only in the sealing
	// step; but a brand-new row cannot arrive already "sealed"
	err := d.Tx(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `insert into hub.orgs (slug, name, mac) values ('nomac', 'x', decode(repeat('ab',32),'hex'))`)
		return err
	})
	if err == nil {
		t.Fatal("a new row arrived pre-sealed")
	}
	// update that forgets to advance version/MAC
	err = d.Tx(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `update hub.orgs set name='renamed' where id=$1`, org)
		return err
	})
	if err == nil {
		t.Fatal("update without a new MAC was accepted")
	}
	// proper update
	err = d.Tx(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `update hub.orgs set name='Renamed', `+integrity.Bump+` where id=$1`, org); err != nil {
			return err
		}
		return s.Seal(ctx, tx, "orgs", "id=$1", org)
	})
	if err != nil {
		t.Fatalf("proper update refused: %v", err)
	}
	// skipping the version bump is refused, and so is a version jump
	for _, q := range []string{`update hub.orgs set name='x', mac=hub.pending_mac() where id=$1`, `update hub.orgs set name='x', ver=ver+5, mac=hub.pending_mac() where id=$1`} {
		if err := d.WithOrg(ctx, org, func(tx pgx.Tx) error { _, err := tx.Exec(ctx, q, org); return err }); err == nil {
			t.Fatalf("accepted: %s", q)
		}
	}
	// a pending row is never left behind
	var n int
	_ = d.Tx(ctx, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `select count(*) from hub.orgs where mac = hub.pending_mac()`).Scan(&n)
	})
	if n != 0 {
		t.Fatalf("%d unsealed rows", n)
	}
}

func TestTenantIsolationByRowLevelSecurity(t *testing.T) {
	_, d, s, _ := setup(t)
	ctx := context.Background()
	a, b := newOrg(t, d, s), newOrg(t, d, s)
	mk := func(org uuid.UUID, name string) {
		id := uuid.Must(uuid.NewV7())
		err := d.WithOrg(ctx, org, func(tx pgx.Tx) error {
			if _, err := tx.Exec(ctx, `insert into hub.org_units (id, org_id, kind, name) values ($1,$2,'branch',$3)`, id, org, name); err != nil {
				return err
			}
			return s.Seal(ctx, tx, "org_units", "id=$1", id)
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	mk(a, "A-branch")
	mk(b, "B-branch")
	count := func(org uuid.UUID) (n int, names []string) {
		_ = d.WithOrg(ctx, org, func(tx pgx.Tx) error {
			rows, _ := tx.Query(ctx, `select name from hub.org_units`) // deliberately no WHERE org_id
			defer rows.Close()
			for rows.Next() {
				var nm string
				_ = rows.Scan(&nm)
				names = append(names, nm)
			}
			return nil
		})
		return len(names), names
	}
	if n, names := count(a); n != 1 || names[0] != "A-branch" {
		t.Fatalf("org A saw %v", names)
	}
	// writing a row for another organization must also fail
	err := d.WithOrg(ctx, a, func(tx pgx.Tx) error {
		id := uuid.Must(uuid.NewV7())
		_, err := tx.Exec(ctx, `insert into hub.org_units (id, org_id, kind, name) values ($1,$2,'branch','x')`, id, b)
		return err
	})
	if err == nil {
		t.Fatal("org A wrote into org B")
	}
}

func TestLedgerChainAndTamperDetection(t *testing.T) {
	cfg, d, s, l := setup(t)
	ctx := context.Background()
	org := newOrg(t, d, s)
	actor := uuid.Must(uuid.NewV7())
	for _, a := range []string{"person.created", "role.changed", "settings.updated"} {
		err := d.WithOrg(ctx, org, func(tx pgx.Tx) error {
			_, err := l.Append(ctx, tx, org, &actor, a, "x", []byte(`{"k":1}`))
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	verify := func() error {
		return d.WithOrg(ctx, org, func(tx pgx.Tx) error { _, err := l.Verify(ctx, tx, org); return err })
	}
	if err := verify(); err != nil {
		t.Fatalf("clean ledger failed verification: %v", err)
	}
	// the runtime role can neither update nor delete ledger rows
	for _, q := range []string{`update hub.ledger set action='x' where org_id=$1`, `delete from hub.ledger where org_id=$1`} {
		if err := d.WithOrg(ctx, org, func(tx pgx.Tx) error { _, err := tx.Exec(ctx, q, org); return err }); err == nil {
			t.Fatalf("ledger accepted: %s", q)
		}
	}
	// Someone with the database superuser password who disables triggers and edits history is still caught.
	root, err := pgx.Connect(ctx, "postgres://hubadmin@/hub?host=/tmp&port=5433")
	if err != nil {
		t.Skip("superuser not reachable for tamper simulation: ", err)
	}
	defer root.Close(ctx)
	_, _ = root.Exec(ctx, "set session_replication_role = replica")
	if _, err := root.Exec(ctx, `update hub.ledger set action='forged' where org_id=$1 and seq=2`, org); err != nil {
		t.Skip("cannot simulate tamper: ", err)
	}
	var br *integrity.Break
	if err := verify(); !errors.As(err, &br) || br.Seq != 2 {
		t.Fatalf("tampered entry not detected at seq 2: %v", err)
	}
	// deleting the last entry (cutting the tail) is caught by the head check
	_, _ = root.Exec(ctx, `update hub.ledger set action='role.changed' where org_id=$1 and seq=2`, org)
	if err := verify(); err != nil {
		t.Fatalf("restored ledger should verify: %v", err)
	}
	_, _ = root.Exec(ctx, `delete from hub.ledger where org_id=$1 and seq=3`, org)
	if err := verify(); !errors.As(err, &br) {
		t.Fatalf("removed tail not detected: %v", err)
	}
	_ = cfg
}

func TestRowMACDetectsOutOfBandEdit(t *testing.T) {
	_, d, s, _ := setup(t)
	ctx := context.Background()
	org := newOrg(t, d, s)
	root, err := pgx.Connect(ctx, "postgres://hubadmin@/hub?host=/tmp&port=5433")
	if err != nil {
		t.Skip("superuser not reachable: ", err)
	}
	defer root.Close(ctx)
	_, _ = root.Exec(ctx, "set session_replication_role = replica")
	if _, err := root.Exec(ctx, `update hub.orgs set plan='enterprise' where id=$1`, org); err != nil {
		t.Skip("cannot simulate tamper: ", err)
	}
	var ver int
	var mac []byte
	var canon string
	read := func() {
		if err := d.Tx(ctx, func(tx pgx.Tx) error {
			return tx.QueryRow(ctx, `select ver, mac, hub.canon(o) from hub.orgs o where id=$1`, org).Scan(&ver, &mac, &canon)
		}); err != nil {
			t.Fatal(err)
		}
	}
	read()
	if s.Verify(mac, "orgs", ver, canon) {
		t.Fatal("an edit made outside the system passed MAC verification")
	}
	_, _ = root.Exec(ctx, `update hub.orgs set plan='business' where id=$1`, org)
	read()
	if !s.Verify(mac, "orgs", ver, canon) {
		t.Fatal("the original values should verify again")
	}
}

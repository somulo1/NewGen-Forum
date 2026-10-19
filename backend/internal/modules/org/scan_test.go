package org_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/e2ee"
	"hub/internal/integrity"
	"hub/internal/testkit"
)

func TestScannerFindsEditsMadeOutsideTheSystem(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	org := uuid.MustParse(e.OrgID(slug))
	o := e.Owner(ownerU)
	a := e.Join(ownerU, slug, "Amy", "staff")
	o.Register(e)
	a.Register(e)
	key, _ := e2ee.NewConversationKey()
	conv := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "forum", "name": "General", "members": []string{a.Person}, "keys": o.WrapFor(e, key, 1, o.Person, a.Person)}, ownerU.Access)["id"].(string)
	o.Send(e, e.OrgID(slug), conv, 1, key, "hello", 201)

	ctx := context.Background()
	audit, err := pgx.Connect(ctx, "postgres://hub_audit:hub_audit_dev_pw@127.0.0.1:5433/hub?sslmode=disable")
	if err != nil {
		t.Skip("audit role not reachable: ", err)
	}
	defer audit.Close(ctx)
	root, err := pgx.Connect(ctx, "postgres://hubadmin@/hub?host=/tmp&port=5433")
	if err != nil {
		t.Skip("superuser not reachable: ", err)
	}
	defer root.Close(ctx)
	sc := &integrity.Scanner{S: integrity.NewSigner(e.A.Cfg.MACKey), L: integrity.NewLedger(e.A.Cfg.LedgerKey), OnlyOrgs: []uuid.UUID{org}}
	scan := func(table string) *integrity.Report {
		rep, err := sc.Run(ctx, audit, table)
		if err != nil {
			t.Fatal(err)
		}
		return rep
	}
	for _, tb := range []string{"people", "conversations", "messages", "roles", "role_rights", "approval_flows"} {
		if rep := scan(tb); !rep.OK() || rep.Rows == 0 {
			t.Fatalf("clean %s: rows=%d problems=%v", tb, rep.Rows, rep.Problems)
		}
	}
	if rep := scan("ledger"); !rep.OK() || rep.Ledgers != 1 {
		t.Fatalf("clean ledger: %v", rep.Problems)
	}
	// the audit role itself cannot change anything
	if _, err := audit.Exec(ctx, `update hub.people set title='CEO'`); err == nil {
		t.Fatal("the read-only role wrote data")
	}
	// someone with database superuser rights turns the triggers off and edits a person, a role right and a message
	_, _ = root.Exec(ctx, "set session_replication_role = replica")
	for _, q := range []string{
		`update hub.people set title='CEO' where id=$1`,
		`update hub.role_rights set allowed = not allowed where org_id=$2 and module='security' and right_name='View logs' and role_id=(select id from hub.roles where org_id=$2 and key='staff')`,
	} {
		if _, err := root.Exec(ctx, q, o.Person, org); err != nil {
			t.Skip("cannot simulate tamper: ", err)
		}
	}
	if _, err := root.Exec(ctx, `update hub.messages set ciphertext = ciphertext || '\x00'::bytea where org_id=$1`, org); err != nil {
		t.Fatal(err)
	}
	for _, tb := range []string{"people", "role_rights", "messages"} {
		rep := scan(tb)
		if rep.OK() {
			t.Fatalf("tampering with %s went unnoticed", tb)
		}
		if rep.Problems[0].Org != org.String() {
			t.Fatalf("problem not attributed to the right organization: %v", rep.Problems[0])
		}
	}
	// and a row inserted behind the system's back (never sealed) is reported too
	if _, err := root.Exec(ctx, `insert into hub.org_units (org_id, kind, name, mac) values ($1,'branch','Ghost', decode(repeat('ab',32),'hex'))`, org); err != nil {
		t.Fatal(err)
	}
	if rep := scan("org_units"); rep.OK() {
		t.Fatal("a forged row went unnoticed")
	}
}

package security_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"

	"hub/internal/integrity"
	"hub/internal/testkit"
)

// A table shared by every organization (account_orgs has an org_id column but is not filtered by organization) is read once, and a
// changed record is reported once, under the organization it names, not once per organization in the system.
func TestSharedTableIsScannedOnceAndBlamedOnItsOwnOrg(t *testing.T) {
	e := testkit.New(t)
	_, slugA := e.NewOrg("Ann Owner")
	ownerB, slugB := e.NewOrg("Bob Owner")
	_ = slugA
	orgB := e.OrgID(slugB)
	acct := e.MustDo(200, "GET", "/me", nil, ownerB.Access)["account"].(string)

	root, err := pgx.Connect(context.Background(), "postgres://hubadmin@/hub?host=/tmp&port=5433")
	if err != nil {
		t.Skip("superuser not reachable: ", err)
	}
	defer root.Close(context.Background())
	// change one membership record outside the system (without going through its seal)
	_, _ = root.Exec(context.Background(), "set session_replication_role = replica")
	if _, err := root.Exec(context.Background(), `update hub.account_orgs set ver = ver + 7 where account_id=$1 and org_id=$2`, acct, orgB); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = root.Exec(context.Background(), `update hub.account_orgs set ver = ver - 7 where account_id=$1 and org_id=$2`, acct, orgB)
	})

	audit, err := pgx.Connect(context.Background(), "postgres://hub_audit:hub_audit_dev_pw@127.0.0.1:5433/hub?sslmode=disable")
	if err != nil {
		t.Skip("audit role not reachable: ", err)
	}
	defer audit.Close(context.Background())
	sc := &integrity.Scanner{S: e.A.Sign, L: e.A.Ledger, MaxBad: 1000}
	rep, err := sc.Run(context.Background(), audit, "account_orgs")
	if err != nil {
		t.Fatal(err)
	}
	var mine []integrity.Problem
	for _, p := range rep.Problems {
		if p.PK["account_id"] == acct {
			mine = append(mine, p)
		}
	}
	if len(mine) != 1 {
		t.Fatalf("the changed record was reported %d times (once per organization used to be the case)", len(mine))
	}
	if mine[0].Org != orgB {
		t.Fatalf("blamed organization %s, the record belongs to %s", mine[0].Org, orgB)
	}
	if rep.Rows > 1_000_000 {
		t.Fatalf("scanned %d rows: the shared table was read once per organization", rep.Rows)
	}
}

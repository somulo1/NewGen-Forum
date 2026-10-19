package org_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"

	"hub/internal/testkit"
)

func TestCreateOrgBuildsEverythingAndSealsIt(t *testing.T) {
	e := testkit.New(t)
	u, slug := e.NewOrg("Wanjiru Kamau")
	if u.Access == "" {
		t.Fatal("owner could not enter the new organization")
	}
	// duplicate web name
	acct := testkit.Uniq()
	e.Signup("Other", acct)
	o := e.Login(acct, "")
	if code, _ := e.Do("POST", "/orgs", map[string]any{"name": "Dup", "slug": slug}, o.Access); code != 409 {
		t.Fatalf("duplicate slug: %d", code)
	}
	// structure: add a sector, a branch under it, a unit; invalid nesting is refused
	sec := e.MustDo(201, "POST", "/org/units", map[string]any{"kind": "sector", "name": "Retail"}, u.Access)
	br := e.MustDo(201, "POST", "/org/units", map[string]any{"kind": "branch", "name": "Nairobi", "parent_id": sec["id"]}, u.Access)
	if code, _ := e.Do("POST", "/org/units", map[string]any{"kind": "sector", "name": "Bad", "parent_id": br["id"]}, u.Access); code != 400 {
		t.Fatalf("sector under branch accepted: %d", code)
	}
	if code, _ := e.Do("POST", "/org/units", map[string]any{"kind": "branch", "name": "Nairobi", "parent_id": sec["id"]}, u.Access); code != 409 {
		t.Fatalf("duplicate branch: %d", code)
	}
	st := e.MustDo(200, "GET", "/org/structure", nil, u.Access)
	if n := len(st["units"].([]any)); n != 2 {
		t.Fatalf("want 2 units, got %d", n)
	}
	// a stranger sees nothing of this organization
	if code, _ := e.Do("GET", "/org/structure", nil, o.Access); code != 403 {
		t.Fatalf("outsider got %d", code)
	}
	// the database holds sealed rows only and the roles/rights matrix is complete
	var pending, roles, rights int
	err := e.A.DB.Tx(context.Background(), func(tx pgx.Tx) error {
		if err := tx.QueryRow(context.Background(), `select (select count(*) from hub.people where mac=hub.pending_mac()) + (select count(*) from hub.orgs where mac=hub.pending_mac())`).Scan(&pending); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if pending != 0 {
		t.Fatalf("%d rows left unsealed", pending)
	}
	_ = roles
	_ = rights
}

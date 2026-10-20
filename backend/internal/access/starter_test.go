package access_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/access"
	"hub/internal/app"
	"hub/internal/testkit"
)

// A real organization starts with only its owner role. The starter pack adds ready-made roles on request, only adds what is
// missing, never overwrites what the organization changed, and holds no demo content.
func TestStarterRolesAreIdempotentAndSeparateFromDemo(t *testing.T) {
	e := testkit.New(t)
	ctx := context.Background()
	// made directly (testkit.NewOrg adds the demo roles, which is what this test must not have)
	email := testkit.Uniq()
	e.Signup("Olivia Owner", email)
	u := e.Login(email, "")
	slug := "t-" + strings.ToLower(strings.ReplaceAll(uuid.NewString(), "-", "")[:14])
	e.MustDo(201, "POST", "/orgs", map[string]any{"name": "Starter Co", "slug": slug}, u.Access)
	org := uuid.MustParse(e.OrgID(slug))
	seal := func(tx pgx.Tx) func(table, where string, args ...any) error {
		return func(table, where string, args ...any) error { return e.A.Sign.Seal(ctx, tx, table, where, args...) }
	}
	count := func(q string) (n int) {
		_ = e.A.DB.WithOrg(ctx, org, func(tx pgx.Tx) error { return tx.QueryRow(ctx, q).Scan(&n) })
		return
	}
	if n := count(`select count(*) from hub.roles`); n != 1 {
		t.Fatalf("a new organization should have only its owner role, has %d", n)
	}
	apply := func() int {
		var added int
		err := e.A.DB.WithOrg(ctx, org, func(tx pgx.Tx) error {
			ids, err := access.InstallStarterRoles(ctx, tx, org, app.NewID, seal(tx))
			added = len(ids)
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
		return added
	}
	if n := apply(); n != 9 {
		t.Fatalf("starter roles added: %d", n)
	}
	if n := count(`select count(*) from hub.roles`); n != 10 {
		t.Fatalf("roles after: %d", n)
	}
	if n := count(`select count(*) from hub.default_role`); n != 1 {
		t.Fatalf("a default role must be set: %d", n)
	}
	if n := count(`select count(*) from hub.approval_flows where levels > 0`); n != 0 {
		t.Fatalf("the starter pack must not turn on approvals: %d", n)
	}
	// the organization changes a role; applying again must not touch it
	_ = e.A.DB.WithOrg(ctx, org, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `delete from hub.role_rights where role_id=(select id from hub.roles where key='guest')`)
		return err
	})
	if n := apply(); n != 0 {
		t.Fatalf("running again should add nothing, added %d", n)
	}
	if n := count(`select count(*) from hub.role_rights where role_id=(select id from hub.roles where key='guest')`); n != 0 {
		t.Fatalf("running again must not restore rights the organization removed: %d", n)
	}
	for _, r := range access.StarterPack().Roles {
		if strings.EqualFold(r.Name, "Super admin") || r.Locked {
			t.Fatalf("the starter pack must not carry the owner role: %s", r.Name)
		}
	}
}

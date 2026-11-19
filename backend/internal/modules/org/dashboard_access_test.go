package org_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/orgsettings"
	"hub/internal/testkit"
	"hub/internal/widgetaccess"
)

type dashCase struct {
	e      *testkit.Env
	owner  string
	amy    *testkit.Member
	dept   string
	org    uuid.UUID
	module string
}

func newDashCase(t *testing.T) *dashCase {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	e.MustDo(204, "PUT", "/approvals/flows/f11", map[string]any{"levels": 0}, ownerU.Access)
	e.MustDo(204, "PUT", "/approvals/flows/f8", map[string]any{"levels": 0}, ownerU.Access)
	e.Levels(ownerU, map[string]string{"department": ""})
	dept := e.MustDo(201, "POST", "/org/units", map[string]any{"kind": "department", "name": "Finance"}, ownerU.Access)["id"].(string)
	amy := e.Join(ownerU, slug, "Amy", "hod")
	e.MustDo(200, "PATCH", "/people/"+amy.Person, map[string]any{"department_id": dept}, ownerU.Access)
	return &dashCase{e: e, owner: ownerU.Access, amy: amy, dept: dept, org: uuid.MustParse(e.OrgID(slug)), module: "departments"}
}

// agree checks that the sidebar's list (GET /me/widgets) and the dashboard itself give the same answer.
func (d *dashCase) agree(t *testing.T, want bool, why string) {
	t.Helper()
	listed := false
	for _, x := range d.e.MustDo(200, "GET", "/me/widgets", nil, d.amy.Access)["allowed"].([]any) {
		listed = listed || x == d.module
	}
	c, _ := d.e.Do("GET", "/org/units/"+d.dept+"/dashboard", nil, d.amy.Access)
	if listed != want || (c == 200) != want {
		t.Fatalf("%s: listed=%v dashboard=%d, want open=%v", why, listed, c, want)
	}
}

func (d *dashCase) right(allowed bool) {
	d.e.MustDo(200, "PUT", "/roles/hod/rights", map[string]any{"module": d.module, "right": "View dashboard", "allowed": allowed}, d.owner)
}

// Who opens a place's dashboard is one right on Roles & rights: the role's "View dashboard", once the widget is switched on.
func TestTheRoleRightOpensADashboard(t *testing.T) {
	d := newDashCase(t)
	d.right(true)
	d.agree(t, false, "widget off")
	d.e.MustDo(204, "PUT", "/settings/widgets", map[string]any{"departments": map[string]any{"enabled": true}}, d.owner)
	d.agree(t, true, "role given View dashboard")
	d.right(false)
	d.agree(t, false, "role's right removed")
}

// Retiring user groups keeps what each role could do: a group's rights land on its roles, and a dashboard that only user groups
// opened keeps exactly the roles the groups let in.
func TestUserGroupsFoldIntoRoles(t *testing.T) {
	d := newDashCase(t)
	ctx := context.Background()
	d.right(true) // the role's own tick did not count for a dashboard opened by user groups only
	if err := d.e.A.DB.WithOrg(ctx, d.org, func(tx pgx.Tx) error {
		return orgsettings.Set(ctx, d.e.A, tx, d.org, "widgets", map[string]any{"departments": map[string]any{"enabled": true, "access": "groups", "see": "Everyone"}})
	}); err != nil {
		t.Fatal(err)
	}
	g := uuid.New()
	d.e.RootExec(`insert into hub.role_groups (id, org_id, key, name) values ($1, $2, 'heads', 'Heads')`, g, d.org)
	d.e.RootExec(`insert into hub.role_group_roles (org_id, group_id, role_id) select $1, $2, id from hub.roles where key='hod'`, d.org, g)
	d.e.RootExec(`insert into hub.role_group_rights (org_id, group_id, module, right_name, allowed) values ($1, $2, 'tickets', 'Raise', true)`, d.org, g)

	widgetaccess.FoldUserGroups(ctx, d.e.A)

	d.agree(t, false, "the group never gave View dashboard")
	rights := d.e.MustDo(200, "GET", "/roles/hod", nil, d.owner)["rights"].(map[string]any)
	if r, _ := rights["tickets"].(map[string]any); r["Raise"] != true {
		t.Fatalf("the group's right did not move onto the role: %v", rights["tickets"])
	}
	var left int
	_ = d.e.A.DB.WithOrg(ctx, d.org, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `select (select count(*) from hub.role_groups) + (select count(*) from hub.role_group_rights)`).Scan(&left)
	})
	if left != 0 {
		t.Fatalf("%d user group rows are left", left)
	}
	if c := d.e.MustDo(200, "GET", "/settings/widgets", nil, d.owner)["departments"].(map[string]any); c["access"] != nil || c["see"] != nil {
		t.Fatalf("old audience fields are still saved: %v", c)
	}
}

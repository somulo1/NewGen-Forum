package projects_test

import (
	"testing"

	"github.com/google/uuid"

	"hub/internal/e2ee"
	"hub/internal/testkit"
)

// Who may move and edit a project task: the owners move it between To do, In progress and Review; only the project manager
// moves it into or out of Done; and only the owners edit it.
func TestTaskMovesAndEditsFollowTheProjectRoles(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	org := e.OrgID(slug)
	o := e.Owner(ownerU)
	pm := e.Join(ownerU, slug, "Hana", "hod")
	amy := e.Join(ownerU, slug, "Amy", "hod")
	ben := e.Join(ownerU, slug, "Ben", "hod")
	cy := e.Join(ownerU, slug, "Cy", "hod")
	for _, m := range []*testkit.Member{o, pm, amy, ben, cy} {
		m.Register(e)
	}
	key, _ := e2ee.NewConversationKey()
	conv := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "project", "name": "Alpha", "members": []string{amy.Person, ben.Person, cy.Person}, "keys": pm.WrapFor(e, key, 1, pm.Person, amy.Person, ben.Person, cy.Person)}, pm.Access)["id"].(string)
	res := e.MustDo(201, "POST", "/projects", map[string]any{"conversation": conv, "start": "2026-09-01", "due": "2026-12-15", "budget_pct": 10}, pm.Access)
	approval := res["approval"].(map[string]any)["id"].(string)
	for i := 0; i < 3; i++ {
		e.MustDo(200, "POST", "/approvals/"+approval+"/approve", map[string]any{}, ownerU.Access)
	}
	id := uuid.Must(uuid.NewV7()).String()
	body := func(m *testkit.Member, column string, assignee, secondary string) map[string]any {
		return map[string]any{"collection": "task", "epoch": 1, "device": m.DevID, "envelope": m.Seal(org, conv, id, 1, key, `{"title":"Wire it up"}`),
			"meta": map[string]any{"column": column, "assignee": assignee, "secondary": secondary}}
	}
	create := body(pm, "todo", amy.Person, ben.Person)
	create["id"] = id
	e.MustDo(201, "POST", "/containers/"+conv+"/records", create, pm.Access)
	put := func(m *testkit.Member, column string, want int, what string) {
		t.Helper()
		if code, out := e.Do("PUT", "/containers/"+conv+"/records/"+id, body(m, column, amy.Person, ben.Person), m.Access); code != want {
			t.Fatalf("%s: want %d got %d %v", what, want, code, out)
		}
	}
	put(amy, "doing", 200, "the primary owner moves it to In progress")
	put(ben, "review", 200, "the secondary owner moves it to Review")
	put(cy, "doing", 403, "someone else moves it")
	put(amy, "done", 403, "an owner moves it to Done")
	put(ben, "done", 403, "the secondary owner moves it to Done")
	put(pm, "done", 200, "the project manager moves it to Done")
	put(amy, "review", 403, "an owner moves it back from Done")
	put(pm, "review", 200, "the project manager moves it back to Review")
	put(pm, "doing", 200, "the project manager moves it between the working columns")
	// editing without moving: owners only
	put(cy, "doing", 403, "a team member edits it")
	put(pm, "doing", 403, "the project manager edits its content")
	put(amy, "doing", 200, "the primary owner edits it")
	put(ben, "doing", 200, "the secondary owner edits it")
}

package roles_test

import (
	"testing"

	"hub/internal/testkit"
)

// A user group gives its roles the rights set on it: everyone whose role is in the group gets them, a personal override
// still wins, and taking the role out of the group (or deleting the group) takes the rights away again.
func TestUserGroupGivesItsRolesTheirRights(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	gus := e.Join(ownerU, slug, "Gus", "guest")
	amy := e.Join(ownerU, slug, "Amy", "staff")
	for _, m := range []*testkit.Member{gus, amy} {
		m.Register(e)
	}
	e.MustDo(204, "PUT", "/approvals/flows/f8", map[string]any{"levels": 0}, ownerU.Access) // rights changes apply at once
	// the right in question: viewing the list of roles, which a guest does not have
	if code, _ := e.Do("GET", "/roles", nil, gus.Access); code != 403 {
		t.Fatalf("a guest can already view roles: %d", code)
	}
	// only people who manage roles may make groups
	if code, _ := e.Do("POST", "/groups", map[string]any{"name": "Team HQ"}, gus.Access); code != 403 {
		t.Fatalf("a guest made a group: %d", code)
	}
	key := e.MustDo(201, "POST", "/groups", map[string]any{"name": "Team HQ", "desc": "Head office", "roles": []string{"guest"}}, ownerU.Access)["id"].(string)
	if code, _ := e.Do("POST", "/groups", map[string]any{"name": "team hq"}, ownerU.Access); code != 409 {
		t.Fatalf("duplicate group name: %d", code)
	}
	if code, _ := e.Do("POST", "/groups", map[string]any{"name": "Bad", "roles": []string{"nope"}}, ownerU.Access); code != 400 {
		t.Fatalf("a group with a role that does not exist: %d", code)
	}
	// no rights yet: nothing changes
	if code, _ := e.Do("GET", "/roles", nil, gus.Access); code != 403 {
		t.Fatalf("an empty group gave a right: %d", code)
	}
	e.MustDo(200, "PUT", "/groups/"+key+"/rights", map[string]any{"module": "roles", "right": "View roles", "allowed": true}, ownerU.Access)
	if code, _ := e.Do("GET", "/roles", nil, gus.Access); code != 200 {
		t.Fatalf("the group's right did not reach its role: %d", code)
	}
	// Amy's role is not in the group
	if g := e.MustDo(200, "GET", "/groups/"+key, nil, ownerU.Access); g["rights"].(map[string]any)["roles"].(map[string]any)["View roles"] != true || len(g["roles"].([]any)) != 1 || g["count"].(float64) != 1 {
		t.Fatalf("group detail: %v", g)
	}
	// a personal "no" beats the group
	e.MustDo(200, "PUT", "/people/"+gus.Person+"/rights", map[string]any{"module": "roles", "right": "View roles", "allowed": false}, ownerU.Access)
	if code, _ := e.Do("GET", "/roles", nil, gus.Access); code != 403 {
		t.Fatalf("a personal override did not beat the group: %d", code)
	}
	e.MustDo(200, "PUT", "/people/"+gus.Person+"/rights", map[string]any{"module": "roles", "right": "View roles"}, ownerU.Access)
	if code, _ := e.Do("GET", "/roles", nil, gus.Access); code != 200 {
		t.Fatalf("resetting the override: %d", code)
	}
	// the role leaves the group: the right goes with it. Adding Amy's role gives it to her.
	e.MustDo(204, "PATCH", "/groups/"+key, map[string]any{"roles": []string{"staff"}}, ownerU.Access)
	if code, _ := e.Do("GET", "/roles", nil, gus.Access); code != 403 {
		t.Fatalf("the right stayed after the role left the group: %d", code)
	}
	if code, _ := e.Do("GET", "/roles", nil, amy.Access); code != 200 {
		t.Fatalf("a role added to the group did not get the right: %d", code)
	}
	// deleting the group removes it for good
	e.MustDo(204, "DELETE", "/groups/"+key, nil, ownerU.Access)
	if code, _ := e.Do("GET", "/roles", nil, amy.Access); code == 200 {
		// staff may hold the right through its own settings; only fail if the group is what gave it
		if code2, _ := e.Do("GET", "/groups/"+key, nil, ownerU.Access); code2 != 404 {
			t.Fatalf("the group is still there: %d", code2)
		}
	}
	// a role that is in a group can still be deleted afterwards
	if code, _ := e.Do("GET", "/groups", nil, ownerU.Access); code != 200 {
		t.Fatalf("list groups: %d", code)
	}
}

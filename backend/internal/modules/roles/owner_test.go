package roles_test

import (
	"testing"

	"hub/internal/testkit"
)

// The owner role's rights are changed like any role's, except the two that keep roles manageable; the role moves from one
// person to another, but the organization always keeps an active owner.
func TestOwnerRightsChangeAndTheRoleIsTransferred(t *testing.T) {
	e := testkit.New(t)
	owner, slug := e.NewOrg("Olivia Owner")
	e.MustDo(204, "PUT", "/approvals/flows/f8", map[string]any{"levels": 0}, owner.Access)
	e.MustDo(204, "PUT", "/approvals/flows/f11", map[string]any{"levels": 0}, owner.Access)
	sam, samID := join(t, e, owner, slug, "Sam", "staff")
	me := e.MustDo(200, "GET", "/me", nil, owner.Access)
	ownerID := me["person"].(string)

	e.MustDo(200, "PUT", "/roles/owner/rights", map[string]any{"module": "wall", "right": "Post", "allowed": false}, owner.Access)
	if me := e.MustDo(200, "GET", "/me", nil, owner.Access); has(me, "wall", "Post") || !has(me, "tickets", "View") {
		t.Fatalf("owner rights after switching off wall Post: %v", me["rights"])
	}
	rights := e.MustDo(200, "GET", "/roles/owner", nil, owner.Access)["rights"].(map[string]any)
	if rights["wall"].(map[string]any)["Post"] != false || rights["tickets"].(map[string]any)["View"] != true {
		t.Fatalf("the owner role's matrix: wall %v tickets %v", rights["wall"], rights["tickets"])
	}
	kept := map[string]any{"changes": []any{map[string]any{"module": "roles", "right": "View roles", "allowed": false}}}
	if c, _ := e.Do("PUT", "/roles/owner/rights/batch", kept, owner.Access); c != 409 {
		t.Fatalf("the owner role lost View roles in a batch: %d", c)
	}

	if c, _ := e.Do("PATCH", "/people/"+ownerID, map[string]any{"role": "staff"}, owner.Access); c != 409 {
		t.Fatalf("the last owner left the role: %d", c)
	}
	if c, _ := e.Do("PATCH", "/people/"+ownerID, map[string]any{"role": "owner"}, sam.Access); c != 403 {
		t.Fatalf("staff gave themselves a say over the owner: %d", c)
	}
	e.MustDo(200, "PATCH", "/people/"+samID, map[string]any{"role": "owner"}, owner.Access)
	e.MustDo(200, "PATCH", "/people/"+ownerID, map[string]any{"role": "staff"}, owner.Access)
	if !has(e.MustDo(200, "GET", "/me", nil, sam.Access), "roles", "Manage roles & rights") {
		t.Fatal("the new owner cannot manage roles")
	}
	if c, _ := e.Do("PUT", "/roles/staff/rights", map[string]any{"module": "wall", "right": "Post", "allowed": true}, owner.Access); c != 403 {
		t.Fatalf("the former owner still manages rights: %d", c)
	}
	if c, _ := e.Do("PATCH", "/people/"+samID, map[string]any{"role": ""}, sam.Access); c != 409 {
		t.Fatalf("the only owner removed their own role: %d", c)
	}
}

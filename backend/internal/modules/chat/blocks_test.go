package chat_test

import (
	"testing"

	"github.com/google/uuid"

	"hub/internal/e2ee"
	"hub/internal/testkit"
)

// What the organization blocks in its settings is refused for everyone, and deleting someone else's post needs its own right.
func TestOrganizationBlocksAndDeletingOthersPosts(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	org := e.OrgID(slug)
	o := e.Owner(ownerU)
	a := e.Join(ownerU, slug, "Amy", "staff")
	for _, m := range []*testkit.Member{o, a} {
		m.Register(e)
	}
	e.MustDo(204, "PUT", "/approvals/flows/f8", map[string]any{"levels": 0}, ownerU.Access)
	key, _ := e2ee.NewConversationKey()
	wall := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "wall", "name": "Wall", "members": []string{a.Person}, "keys": o.WrapFor(e, key, 1, o.Person, a.Person)}, ownerU.Access)["id"].(string)
	post := func(m *testkit.Member, access string) (string, int) {
		id := uuid.Must(uuid.NewV7()).String()
		c, _ := e.Do("POST", "/conversations/"+wall+"/messages", map[string]any{"id": id, "epoch": 1, "device": m.DevID, "type": "post", "envelope": m.Seal(org, wall, id, 1, key, "hello")}, access)
		return id, c
	}
	id, c := post(o, ownerU.Access)
	if c != 201 {
		t.Fatalf("post: %d", c)
	}
	if c, _ := e.Do("DELETE", "/conversations/"+wall+"/messages/"+id, nil, a.Access); c != 403 {
		t.Fatalf("deleted someone else's post without the right: %d", c)
	}
	e.MustDo(200, "PUT", "/people/"+a.Person+"/rights", map[string]any{"module": "wall", "right": "Delete others' posts", "allowed": true}, ownerU.Access)
	e.MustDo(204, "DELETE", "/conversations/"+wall+"/messages/"+id, nil, a.Access)

	sec := e.MustDo(200, "GET", "/settings", nil, ownerU.Access)["security"].(map[string]any)
	sec["blockWall"], sec["blockPrivateVoice"], sec["blockForumVideo"] = true, true, true
	e.MustDo(204, "PUT", "/settings/security", sec, ownerU.Access)
	if b := e.MustDo(200, "GET", "/settings/blocks", nil, a.Access); b["blockWall"] != true || b["blockForumVoice"] != false {
		t.Fatalf("blocks: %v", b)
	}
	if _, c := post(o, ownerU.Access); c != 403 {
		t.Fatalf("posted on a blocked wall: %d", c)
	}
	if c, _ := e.Do("GET", "/conversations/"+wall+"/messages", nil, a.Access); c != 403 {
		t.Fatalf("read a blocked wall: %d", c)
	}
	if c, _ := e.Do("POST", "/calls", map[string]any{"kind": "voice", "context": "dm"}, ownerU.Access); c != 403 {
		t.Fatalf("private voice call: %d", c)
	}
	if c, _ := e.Do("POST", "/calls", map[string]any{"kind": "video", "context": "forum"}, ownerU.Access); c != 403 {
		t.Fatalf("forum video call: %d", c)
	}
	sec["blockPrivateChats"] = true
	e.MustDo(204, "PUT", "/settings/security", sec, ownerU.Access)
	dm := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "dm", "members": []string{a.Person}, "keys": o.WrapFor(e, key, 1, o.Person, a.Person)}, ownerU.Access)["id"].(string)
	mid := uuid.Must(uuid.NewV7()).String()
	if c, _ := e.Do("POST", "/conversations/"+dm+"/messages", map[string]any{"id": mid, "epoch": 1, "device": o.DevID, "type": "text", "envelope": o.Seal(org, dm, mid, 1, key, "hi")}, ownerU.Access); c != 403 {
		t.Fatalf("sent a private message: %d", c)
	}
	if c, _ := e.Do("GET", "/conversations/"+dm+"/messages", nil, a.Access); c != 403 {
		t.Fatalf("read a private chat: %d", c)
	}
	// unblocking brings everything back
	for _, k := range []string{"blockWall", "blockPrivateChats"} {
		sec[k] = false
	}
	e.MustDo(204, "PUT", "/settings/security", sec, ownerU.Access)
	if _, c := post(o, ownerU.Access); c != 201 {
		t.Fatalf("post after unblocking: %d", c)
	}
}

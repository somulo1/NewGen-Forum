package chat_test

import (
	"encoding/base64"
	"testing"

	"github.com/google/uuid"

	"hub/internal/e2ee"
	"hub/internal/testkit"
)

func (m *member) record(t *testing.T, e *testkit.Env, org, container string, epoch int, key []byte, collection, text string, meta map[string]any, want int) map[string]any {
	t.Helper()
	id := uuid.Must(uuid.NewV7())
	env, err := m.dev.Seal(key, []byte(text), e2ee.Context{Org: org, Conversation: container, Sender: m.acct, MessageID: id.String(), Epoch: epoch})
	if err != nil {
		t.Fatal(err)
	}
	code, out := e.Do("POST", "/containers/"+container+"/records", map[string]any{"id": id, "collection": collection, "epoch": epoch, "device": m.devID, "envelope": b64(env), "meta": meta}, m.Access)
	if code != want {
		t.Fatalf("record %s: want %d got %d %v", collection, want, code, out)
	}
	return out
}

func TestNotesVaultIsPrivateAndProjectsEnforceRights(t *testing.T) {
	e := testkit.New(t)
	owner, slug := e.NewOrg("Olivia Owner")
	org := orgID(t, e, slug)
	o := &member{User: owner}
	o.person = e.MustDo(200, "GET", "/me", nil, owner.Access)["person"].(string)
	a := newMember(t, e, owner, slug, "Amy")
	o.register(t, e)
	a.register(t, e)

	// Amy's private notes vault: one per person, only she can open it
	key, _ := e2ee.NewConversationKey()
	body := map[string]any{"kind": "note", "keys": wrapFor(t, e, a, key, 1, "", a.person)}
	v := e.MustDo(201, "POST", "/conversations", body, a.Access)
	if again := e.MustDo(200, "POST", "/conversations", body, a.Access); again["id"] != v["id"] {
		t.Fatal("second notes vault created")
	}
	vid := v["id"].(string)
	n := a.record(t, e, org, vid, 1, key, "note", "buy milk; fire Bob", map[string]any{"pinned": true}, 201)
	if code, _ := e.Do("GET", "/containers/"+vid+"/records", nil, owner.Access); code != 404 {
		t.Fatalf("even the super admin read someone's private notes: %d", code)
	}
	a.record(t, e, org, vid, 1, key, "note", "x", map[string]any{"secret": "leak"}, 400)
	a.record(t, e, org, vid, 1, key, "task", "x", nil, 400) // wrong container kind
	nid := n["id"].(string)
	env, _ := e2ee.NewConversationKey()
	_ = env
	sealed, _ := a.dev.Seal(key, []byte("buy oat milk"), e2ee.Context{Org: org, Conversation: vid, Sender: a.acct, MessageID: nid, Epoch: 1})
	e.MustDo(200, "PUT", "/containers/"+vid+"/records/"+nid, map[string]any{"collection": "note", "epoch": 1, "device": a.devID, "envelope": b64(sealed), "meta": map[string]any{}}, a.Access)
	got := e.MustDo(200, "GET", "/containers/"+vid+"/records?collection=note", nil, a.Access)["records"].([]any)
	if len(got) != 1 {
		t.Fatalf("notes: %v", got)
	}
	e.MustDo(204, "DELETE", "/containers/"+vid+"/records/"+nid, nil, a.Access)
	got = e.MustDo(200, "GET", "/containers/"+vid+"/records", nil, a.Access)["records"].([]any)
	if len(got) != 1 || got[0].(map[string]any)["deleted"] != true || got[0].(map[string]any)["envelope"] != nil {
		t.Fatalf("deleted note still readable: %v", got)
	}

	// a project: Amy (staff) may edit tasks but not manage milestones
	pk, _ := e2ee.NewConversationKey()
	proj := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "project", "name": "Mobile App 2.0", "members": []string{a.person}, "keys": wrapFor(t, e, o, pk, 1, "", o.person, a.person)}, owner.Access)
	pid := proj["id"].(string)
	a.record(t, e, org, pid, 1, pk, "task", "design login", map[string]any{"column": "todo", "due": "2026-11-01"}, 201)
	a.record(t, e, org, pid, 1, pk, "task", "bad", map[string]any{"column": "someday"}, 400)
	a.record(t, e, org, pid, 1, pk, "task", "bad", map[string]any{"due": "tomorrow"}, 400)
	a.record(t, e, org, pid, 1, pk, "milestone", "beta", map[string]any{"due": "2026-12-01"}, 403)
	o.record(t, e, org, pid, 1, pk, "milestone", "beta", map[string]any{"due": "2026-12-01"}, 201)
	// staff may not create projects
	if code, _ := e.Do("POST", "/conversations", map[string]any{"kind": "project", "name": "Nope", "keys": wrapFor(t, e, a, pk, 1, "", a.person)}, a.Access); code != 403 {
		t.Fatalf("staff created a project: %d", code)
	}
}

func TestWallAudienceKeysFollowMembership(t *testing.T) {
	e := testkit.New(t)
	owner, slug := e.NewOrg("Olivia Owner")
	org := orgID(t, e, slug)
	o := &member{User: owner}
	o.person = e.MustDo(200, "GET", "/me", nil, owner.Access)["person"].(string)
	a := newMember(t, e, owner, slug, "Amy")
	o.register(t, e)
	a.register(t, e)

	k1, _ := e2ee.NewConversationKey()
	wall := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "wall", "name": "Everyone", "audience": "org", "keys": wrapFor(t, e, o, k1, 1, "", o.person, a.person)}, owner.Access)
	wid := wall["id"].(string)
	if int(wall["member_count"].(float64)) != 2 {
		t.Fatalf("audience members: %v", wall["member_count"])
	}
	// staff may post and comment; reading is by membership
	o.send(t, e, org, wid, 1, k1, "welcome", 201)
	if code, _ := e.Do("POST", "/conversations", map[string]any{"kind": "wall", "name": "x", "audience": "org", "keys": []any{}}, a.Access); code != 403 {
		t.Fatalf("staff created a wall: %d", code)
	}

	// Ben joins the organization later: he is in the wall's audience at once but his device has no key yet
	b := newMember(t, e, owner, slug, "Ben")
	b.register(t, e)
	todo := e.MustDo(200, "GET", "/e2ee/todo", nil, o.Access)["todo"].([]any)
	if len(todo) != 1 || int(todo[0].(map[string]any)["missing_devices"].(float64)) != 1 {
		t.Fatalf("expected one conversation waiting for a key: %v", todo)
	}
	if code, _ := e.Do("GET", "/conversations/"+wid+"/keys", nil, b.Access); code != 200 {
		t.Fatalf("new audience member cannot ask for keys: %d", code)
	}
	if ks := e.MustDo(200, "GET", "/conversations/"+wid+"/keys", nil, b.Access)["keys"].([]any); len(ks) != 0 {
		t.Fatalf("keys before anyone shared them: %v", ks)
	}
	missing := e.MustDo(200, "GET", "/conversations/"+wid+"/missing-keys", nil, a.Access)["missing"].([]any)
	if len(missing) != 1 {
		t.Fatalf("missing: %v", missing)
	}
	// Amy's device shares the key it holds
	ak := a.readKey(t, e, wid, 1)
	e.MustDo(200, "POST", "/conversations/"+wid+"/keys", map[string]any{"device": a.devID, "keys": wrapFor(t, e, a, ak, 1, wid, b.person)}, a.Access)
	if bk := b.readKey(t, e, wid, 1); !equal(bk, k1) {
		t.Fatal("Ben could not get the wall key")
	}
	// a stranger's device cannot be handed the key
	z := newMember(t, e, owner, slug, "Zed")
	z.register(t, e)
	stray := wrapFor(t, e, a, ak, 1, wid, z.person)
	stray[0]["device"] = uuid.NewString()
	if code, _ := e.Do("POST", "/conversations/"+wid+"/keys", map[string]any{"device": a.devID, "keys": stray}, a.Access); code != 400 {
		t.Fatalf("key given to a device that is not in the wall: %d", code)
	}
	// Zed is in the whole-organization audience automatically, so a key for his device is fine
	e.MustDo(200, "POST", "/conversations/"+wid+"/keys", map[string]any{"device": a.devID, "keys": wrapFor(t, e, a, ak, 1, wid, z.person)}, a.Access)
	_ = z

	// Ben is blocked: he leaves every wall and the wall needs a new key
	e.MustDo(204, "POST", "/people/"+b.person+"/status", map[string]any{"action": "block"}, owner.Access)
	if code, _ := e.Do("GET", "/conversations/"+wid+"/messages", nil, b.Access); code != 401 && code != 404 {
		t.Fatalf("blocked person still reads the wall: %d", code)
	}
	todo = e.MustDo(200, "GET", "/e2ee/todo", nil, a.Access)["todo"].([]any)
	rot := false
	for _, x := range todo {
		rot = rot || x.(map[string]any)["rotate"] == true
	}
	if !rot {
		t.Fatalf("no rotation requested after a person left: %v", todo)
	}
	k2, _ := e2ee.NewConversationKey()
	// Zed joined too, so he must be covered by the new key as well
	e.MustDo(204, "POST", "/conversations/"+wid+"/rotate", map[string]any{"device": a.devID, "keys": wrapFor(t, e, a, k2, 2, wid, a.person, o.person, z.person)}, a.Access)
	o.send(t, e, org, wid, 2, k2, "after the rotation", 201)
}

func equal(a, b []byte) bool {
	return base64.StdEncoding.EncodeToString(a) == base64.StdEncoding.EncodeToString(b)
}

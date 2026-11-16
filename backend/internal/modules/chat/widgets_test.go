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
	// the unsaved text of the Home page's "My notes" box is a private draft in the same vault, and can be taken away again
	dr := a.record(t, e, org, vid, 1, key, "draft", "half a thought", map[string]any{}, 201)
	e.MustDo(204, "DELETE", "/containers/"+vid+"/records/"+dr["id"].(string), nil, a.Access)

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

	// Ben joins the organization later: he is in the wall's audience at once. The wall is server-side (not end-to-end
	// encrypted), so nobody's device has key work to do for it, however many people join or leave.
	b := newMember(t, e, owner, slug, "Ben")
	b.register(t, e)
	for _, who := range []*member{o, a, b} {
		for _, x := range e.MustDo(200, "GET", "/e2ee/todo", nil, who.Access)["todo"].([]any) {
			if x.(map[string]any)["conversation"] == wid {
				t.Fatalf("the wall asked devices for key work: %v", x)
			}
		}
	}
	if code, _ := e.Do("GET", "/conversations/"+wid+"/messages", nil, b.Access); code != 200 {
		t.Fatalf("a new member cannot read the wall: %d", code)
	}

	// Ben is blocked: he leaves every wall at once
	e.MustDo(204, "POST", "/people/"+b.person+"/status", map[string]any{"action": "block"}, owner.Access)
	if code, _ := e.Do("GET", "/conversations/"+wid+"/messages", nil, b.Access); code != 401 && code != 404 {
		t.Fatalf("blocked person still reads the wall: %d", code)
	}
	_ = k1
}

func equal(a, b []byte) bool {
	return base64.StdEncoding.EncodeToString(a) == base64.StdEncoding.EncodeToString(b)
}

func TestWallIsServerSide(t *testing.T) {
	e := testkit.New(t)
	owner, slug := e.NewOrg("Olivia Owner")
	o := &member{User: owner}
	o.person = e.MustDo(200, "GET", "/me", nil, owner.Access)["person"].(string)
	a := newMember(t, e, owner, slug, "Amy")
	b := newMember(t, e, owner, slug, "Ben")

	// a wall is made with no keys at all
	wid := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "wall", "name": "Everyone", "audience": "org"}, owner.Access)["id"].(string)
	post := func(who *member, typ string, body map[string]any, reply string, want int) string {
		id, _ := uuid.NewV7()
		in := map[string]any{"id": id.String(), "type": typ, "body": body}
		if reply != "" {
			in["reply_to"] = reply
		}
		if code, res := e.Do("POST", "/conversations/"+wid+"/messages", in, who.Access); code != want {
			t.Fatalf("%s %s: want %d got %d %v", who.person, typ, want, code, res)
		}
		return id.String()
	}
	pid := post(o, "post", map[string]any{"text": "Lunch vote", "poll": []any{map[string]any{"id": "o0", "label": "Rice"}, map[string]any{"id": "o1", "label": "Chapati"}}}, "", 201)
	post(a, "comment", map[string]any{"text": "Chapati!"}, pid, 201)
	// likes and votes are not messages any more
	post(a, "reaction", map[string]any{"kind": "like"}, pid, 400)

	// one reaction per person: a new kind replaces the old, the same again is kept, "" takes it back
	e.MustDo(204, "PUT", "/wall/"+pid+"/reaction", map[string]any{"kind": "like"}, a.Access)
	e.MustDo(204, "PUT", "/wall/"+pid+"/reaction", map[string]any{"kind": "like"}, a.Access)
	e.MustDo(204, "PUT", "/wall/"+pid+"/reaction", map[string]any{"kind": "like"}, b.Access)
	e.MustDo(204, "PUT", "/wall/"+pid+"/reaction", map[string]any{"kind": "love"}, b.Access)
	if code, _ := e.Do("PUT", "/wall/"+pid+"/reaction", map[string]any{"kind": "angry"}, b.Access); code != 400 {
		t.Fatalf("unknown reaction: %d", code)
	}
	// a vote is final and must be one of the choices
	e.MustDo(204, "PUT", "/wall/"+pid+"/vote", map[string]any{"option": "o1"}, a.Access)
	if code, _ := e.Do("PUT", "/wall/"+pid+"/vote", map[string]any{"option": "o0"}, a.Access); code != 409 {
		t.Fatalf("second vote: %d", code)
	}
	if code, _ := e.Do("PUT", "/wall/"+pid+"/vote", map[string]any{"option": "o9"}, b.Access); code != 400 {
		t.Fatalf("vote for a choice that does not exist: %d", code)
	}

	find := func(who *member) map[string]any {
		for _, x := range e.MustDo(200, "GET", "/conversations/"+wid+"/messages", nil, who.Access)["messages"].([]any) {
			if m := x.(map[string]any); m["id"] == pid {
				return m
			}
		}
		t.Fatal("post not on the wall")
		return nil
	}
	m := find(a)
	if body := m["body"].(map[string]any); body["text"] != "Lunch vote" || m["envelope"] != "" && m["envelope"] != nil {
		t.Fatalf("the post is not server-readable: %v", m)
	}
	c := m["counts"].(map[string]any)
	if c["like"] != float64(1) || c["love"] != float64(1) || c["vote:o1"] != float64(1) || c["vote:o0"] != nil {
		t.Fatalf("counts: %v", c)
	}
	if m["my_reaction"] != "like" || m["my_vote"] != "o1" {
		t.Fatalf("Amy's own reaction and vote: %v %v", m["my_reaction"], m["my_vote"])
	}
	if mb := find(b); mb["my_reaction"] != "love" || mb["my_vote"] != nil {
		t.Fatalf("Ben's own: %v %v", mb["my_reaction"], mb["my_vote"])
	}
	// taking it back
	e.MustDo(204, "PUT", "/wall/"+pid+"/reaction", map[string]any{"kind": ""}, a.Access)
	if c := find(o)["counts"].(map[string]any); c["like"] != nil {
		t.Fatalf("like still counted after it was taken back: %v", c)
	}
	// a single item's counts, for live updates
	items := e.MustDo(200, "GET", "/wall/counts?ids="+pid, nil, o.Access)["items"].([]any)
	if len(items) != 1 || items[0].(map[string]any)["counts"].(map[string]any)["love"] != float64(1) {
		t.Fatalf("item counts: %v", items)
	}
	// deleting the post takes its likes, votes and counts with it
	e.MustDo(204, "DELETE", "/conversations/"+wid+"/messages/"+pid, nil, owner.Access)
	if items := e.MustDo(200, "GET", "/wall/counts?ids="+pid, nil, o.Access)["items"].([]any); len(items) != 0 {
		t.Fatalf("counts of a deleted post: %v", items)
	}
	if code, _ := e.Do("PUT", "/wall/"+pid+"/reaction", map[string]any{"kind": "like"}, b.Access); code != 404 {
		t.Fatalf("liked a deleted post: %d", code)
	}
}

package chat_test

import (
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"

	"hub/internal/e2ee"
	"hub/internal/testkit"
)

// idAt is a UUID v7 whose time is t, as a phone makes it when the message is written.
func idAt(t time.Time) uuid.UUID {
	id := uuid.Must(uuid.NewV7())
	ms := t.UnixMilli()
	for i := 0; i < 6; i++ {
		id[i] = byte(ms >> (8 * (5 - i)))
	}
	return id
}

func (m *member) sendWithID(t *testing.T, e *testkit.Env, org, conv string, key []byte, id uuid.UUID) int {
	t.Helper()
	env, err := m.dev.Seal(key, []byte("written offline"), e2ee.Context{Org: org, Conversation: conv, Sender: m.acct, MessageID: id.String(), Epoch: 1})
	if err != nil {
		t.Fatal(err)
	}
	code, _ := e.Do("POST", "/conversations/"+conv+"/messages", map[string]any{"id": id, "epoch": 1, "device": m.devID, "envelope": b64(env)}, m.Access)
	return code
}

func TestMessageWrittenOfflineKeepsItsPlace(t *testing.T) {
	e := testkit.New(t)
	owner, slug := e.NewOrg("Olivia Owner")
	org := orgID(t, e, slug)
	o := &member{User: owner}
	o.person = e.MustDo(200, "GET", "/me", nil, owner.Access)["person"].(string)
	a := newMember(t, e, owner, slug, "Amy")
	o.register(t, e)
	a.register(t, e)
	key, _ := e2ee.NewConversationKey()
	cid := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "forum", "name": "Field team", "members": []string{a.person}, "keys": wrapFor(t, e, o, key, 1, "", o.person, a.person)}, owner.Access)["id"].(string)

	before := time.Now().Add(-time.Second).UTC().Format(time.RFC3339Nano)
	o.send(t, e, org, cid, 1, key, "sent online", 201)
	written := idAt(time.Now().Add(-3 * time.Hour))
	if code := o.sendWithID(t, e, org, cid, key, written); code != 201 {
		t.Fatalf("a message written offline 3 hours ago was refused: %d", code)
	}
	// sending it again after a dropped answer is refused as a duplicate, so it never shows twice
	if code := o.sendWithID(t, e, org, cid, key, written); code != 409 {
		t.Fatalf("resending the same offline message: want 409, got %d", code)
	}

	// a colleague who caught up just before it arrived still receives it: catch-up follows the arrival time, not the id
	page := e.MustDo(200, "GET", "/conversations/"+cid+"/changes?since="+url.QueryEscape(before)+"&limit=200", nil, a.Access)
	got := false
	for _, m := range page["messages"].([]any) {
		got = got || m.(map[string]any)["id"] == written.String()
	}
	if !got {
		t.Fatalf("the offline message is missing from catch-up: %v", page)
	}
	// ... and the history (newest first) puts it where it was written: before the message sent online
	hist := e.MustDo(200, "GET", "/conversations/"+cid+"/messages", nil, a.Access)["messages"].([]any)
	if len(hist) != 2 || hist[1].(map[string]any)["id"] != written.String() {
		t.Fatalf("history: %v", hist)
	}

	if code := o.sendWithID(t, e, org, cid, key, idAt(time.Now().Add(-49*time.Hour))); code != 410 {
		t.Fatalf("a message older than 48 hours offline: want 410, got %d", code)
	}
	if code := o.sendWithID(t, e, org, cid, key, idAt(time.Now().Add(10*time.Minute))); code != 400 {
		t.Fatalf("a message dated in the future: want 400, got %d", code)
	}
}

package chat_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"hub/internal/e2ee"
	"hub/internal/testkit"
)

// Deleting a forum or chat for myself empties my copy for good and takes it off my list; it comes back, holding only what is
// written after that, when someone writes again. Everyone else keeps everything.
func TestDeletingAChatForMeOnlyHidesItForMe(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	org := e.OrgID(slug)
	o := e.Owner(ownerU)
	a := e.Join(ownerU, slug, "Amy", "staff")
	for _, m := range []*testkit.Member{o, a} {
		m.Register(e)
	}
	key, _ := e2ee.NewConversationKey()
	conv := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "forum", "name": "Team", "members": []string{a.Person}, "keys": o.WrapFor(e, key, 1, o.Person, a.Person)}, ownerU.Access)["id"].(string)
	say := func(m *testkit.Member, text string) {
		t.Helper()
		id := uuid.Must(uuid.NewV7()).String()
		e.MustDo(201, "POST", "/conversations/"+conv+"/messages", map[string]any{"id": id, "epoch": 1, "device": m.DevID, "envelope": m.Seal(org, conv, id, 1, key, text), "type": "text"}, m.Access)
	}
	count := func(m *testkit.Member) int {
		return len(e.MustDo(200, "GET", "/conversations/"+conv+"/messages?limit=100", nil, m.Access)["messages"].([]any))
	}
	listed := func(m *testkit.Member) bool {
		for _, c := range e.MustDo(200, "GET", "/conversations?kind=forum&limit=200", nil, m.Access)["conversations"].([]any) {
			if c.(map[string]any)["id"] == conv {
				return true
			}
		}
		return false
	}
	say(o, `{"text":"one"}`)
	say(o, `{"text":"two"}`)
	if count(a) != 2 || !listed(a) {
		t.Fatalf("before: %d messages, listed %v", count(a), listed(a))
	}
	time.Sleep(20 * time.Millisecond)
	e.MustDo(204, "POST", "/conversations/"+conv+"/clear", nil, a.Access)
	if ids := e.MustDo(200, "GET", "/conversations/cleared", nil, a.Access)["ids"].([]any); len(ids) != 1 || ids[0] != conv {
		t.Fatalf("cleared list: %v", ids)
	}
	// gone for Amy: not listed, nothing in the history, nothing to catch up on; Olivia still has it all
	if listed(a) || count(a) != 0 {
		t.Fatalf("after deleting: listed %v, %d messages", listed(a), count(a))
	}
	ch := e.MustDo(200, "GET", "/conversations/"+conv+"/changes?since=2000-01-01T00:00:00Z", nil, a.Access)["messages"].([]any)
	if len(ch) != 0 {
		t.Fatalf("changes still carry the deleted history: %d", len(ch))
	}
	if count(o) != 2 || !listed(o) {
		t.Fatalf("Olivia lost hers: %d", count(o))
	}
	// someone writes again: it is back for Amy with only the new message
	time.Sleep(20 * time.Millisecond)
	say(o, `{"text":"three"}`)
	if !listed(a) || count(a) != 1 {
		t.Fatalf("after a new message: listed %v, %d messages", listed(a), count(a))
	}
	if n := e.MustDo(200, "GET", "/conversations?kind=forum&limit=200", nil, a.Access)["conversations"].([]any)[0].(map[string]any)["unread"].(float64); n != 1 {
		t.Fatalf("unread after deleting: %v", n)
	}
	// a person outside it cannot do this to it, and a ticket is not a chat
	out := e.Join(ownerU, slug, "Outsider", "staff")
	out.Register(e)
	if code, _ := e.Do("POST", "/conversations/"+conv+"/clear", nil, out.Access); code != 404 {
		t.Fatalf("a non-member cleared it: %d", code)
	}
}

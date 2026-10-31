package chat_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"hub/internal/e2ee"
	"hub/internal/testkit"
)

// Someone added to a forum later sees only what is written after they join: not in the history, not in what they catch up on,
// and not in their unread count. The people who were there from the start keep everything.
func TestSomeoneAddedLaterSeesNothingFromBefore(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	org := e.OrgID(slug)
	o := e.Owner(ownerU)
	a := e.Join(ownerU, slug, "Amy", "staff")
	for _, m := range []*testkit.Member{o, a} {
		m.Register(e)
	}
	key, _ := e2ee.NewConversationKey()
	conv := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "forum", "name": "Team", "keys": o.WrapFor(e, key, 1, o.Person)}, ownerU.Access)["id"].(string)
	say := func(m *testkit.Member, text string) {
		t.Helper()
		id := uuid.Must(uuid.NewV7()).String()
		e.MustDo(201, "POST", "/conversations/"+conv+"/messages", map[string]any{"id": id, "epoch": 1, "device": m.DevID, "envelope": m.Seal(org, conv, id, 1, key, text), "type": "text"}, m.Access)
	}
	count := func(m *testkit.Member) int {
		return len(e.MustDo(200, "GET", "/conversations/"+conv+"/messages?limit=100", nil, m.Access)["messages"].([]any))
	}
	say(o, `{"text":"one"}`)
	say(o, `{"text":"two"}`)
	time.Sleep(20 * time.Millisecond)
	e.MustDo(200, "POST", "/conversations/"+conv+"/members", map[string]any{"device": o.DevID, "people": []string{a.Person}, "keys": o.WrapFor(e, key, 1, a.Person)}, ownerU.Access)
	time.Sleep(20 * time.Millisecond)
	if n := count(a); n != 0 {
		t.Fatalf("a new member sees %d earlier messages", n)
	}
	if ch := e.MustDo(200, "GET", "/conversations/"+conv+"/changes?since=2000-01-01T00:00:00Z", nil, a.Access)["messages"].([]any); len(ch) != 0 {
		t.Fatalf("catch-up carries %d earlier messages", len(ch))
	}
	unread := func() float64 {
		for _, c := range e.MustDo(200, "GET", "/conversations?kind=forum&limit=200", nil, a.Access)["conversations"].([]any) {
			if c.(map[string]any)["id"] == conv {
				return c.(map[string]any)["unread"].(float64)
			}
		}
		t.Fatal("forum not listed for the new member")
		return 0
	}
	if n := unread(); n != 0 {
		t.Fatalf("unread counts earlier messages: %v", n)
	}
	say(o, `{"text":"three"}`)
	if count(a) != 1 || unread() != 1 {
		t.Fatalf("after joining: %d messages, %v unread", count(a), unread())
	}
	if count(o) != 3 {
		t.Fatalf("the original member lost history: %d", count(o))
	}
}

package chat_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"hub/internal/e2ee"
	"hub/internal/testkit"
)

// Read ticks: each member's reading place is theirs alone, only moves forward, only points at a message of that conversation, and
// is shown to the members of the conversation (not to anyone else).
func TestReadingPlacesAreHonest(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	org := e.OrgID(slug)
	o := e.Owner(ownerU)
	a := e.Join(ownerU, slug, "Amy", "staff")
	b := e.Join(ownerU, slug, "Ben", "staff")
	outsider := e.Join(ownerU, slug, "Zed", "staff")
	for _, m := range []*testkit.Member{o, a, b, outsider} {
		m.Register(e)
	}
	key, _ := e2ee.NewConversationKey()
	conv := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "forum", "name": "Team", "keys": o.WrapFor(e, key, 1, o.Person)}, ownerU.Access)["id"].(string)
	for _, m := range []*testkit.Member{a, b} {
		e.MustDo(200, "POST", "/conversations/"+conv+"/members", map[string]any{"device": o.DevID, "people": []string{m.Person}, "keys": o.WrapFor(e, key, 1, m.Person)}, ownerU.Access)
	}
	say := func(m *testkit.Member, text string) string {
		t.Helper()
		id := uuid.Must(uuid.NewV7()).String()
		e.MustDo(201, "POST", "/conversations/"+conv+"/messages", map[string]any{"id": id, "epoch": 1, "device": m.DevID, "envelope": m.Seal(org, conv, id, 1, key, text), "type": "text"}, m.Access)
		time.Sleep(15 * time.Millisecond)
		return id
	}
	readers := func(who *testkit.Member) map[string]any {
		t.Helper()
		out := map[string]any{}
		for _, r := range e.MustDo(200, "GET", "/conversations/"+conv+"/reads", nil, who.Access)["readers"].([]any) {
			x := r.(map[string]any)
			out[x["person"].(string)] = x["at"]
		}
		return out
	}
	first := say(o, `{"text":"one"}`)
	second := say(o, `{"text":"two"}`)
	got := readers(o)
	if len(got) != 3 || got[a.Person] != nil || got[b.Person] != nil {
		t.Fatalf("before anyone reads: %v", got)
	}
	e.MustDo(204, "POST", "/conversations/"+conv+"/read", map[string]any{"message": second}, a.Access)
	e.MustDo(204, "POST", "/conversations/"+conv+"/read", map[string]any{"message": first}, a.Access) // never back
	got = readers(o)
	if got[a.Person] == nil || got[b.Person] != nil || got[o.Person] != nil {
		t.Fatalf("only Amy has read: %v", got)
	}
	atA := got[a.Person]
	other := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "forum", "name": "Elsewhere", "keys": o.WrapFor(e, key, 1, o.Person)}, ownerU.Access)["id"].(string)
	id := uuid.Must(uuid.NewV7()).String()
	e.MustDo(201, "POST", "/conversations/"+other+"/messages", map[string]any{"id": id, "epoch": 1, "device": o.DevID, "envelope": o.Seal(org, other, id, 1, key, `{"text":"x"}`), "type": "text"}, o.Access)
	e.MustDo(204, "POST", "/conversations/"+conv+"/read", map[string]any{"message": id}, b.Access) // a message of another conversation
	if got = readers(o); got[b.Person] != nil || got[a.Person] != atA {
		t.Fatalf("a foreign message moved a reading place: %v", got)
	}
	if code, _ := e.Do("GET", "/conversations/"+conv+"/reads", nil, outsider.Access); code != 404 {
		t.Fatalf("someone outside the forum saw who has read: %d", code)
	}
}

package chat_test

import (
	"testing"

	"github.com/google/uuid"

	"hub/internal/e2ee"
	"hub/internal/testkit"
)

// "Own" rights are given to roles by the organization: with "Delete own messages" a person can remove what they wrote, with
// "Edit own messages" they can post an edit of it, and neither lets them touch anyone else's messages.
func TestOwnRightsDecideWhatYouMayChange(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	org := e.OrgID(slug)
	e.MustDo(204, "PUT", "/approvals/flows/f8", map[string]any{"levels": 0}, ownerU.Access)
	e.MustDo(204, "PUT", "/approvals/flows/f11", map[string]any{"levels": 0}, ownerU.Access)
	grant := func(role, mod, right string, allowed bool) {
		e.MustDo(200, "PUT", "/roles/"+role+"/rights", map[string]any{"module": mod, "right": right, "allowed": allowed}, ownerU.Access)
	}
	plain := e.MustDo(201, "POST", "/roles", map[string]any{"name": "Reader"}, ownerU.Access)["id"].(string)
	writer := e.MustDo(201, "POST", "/roles", map[string]any{"name": "Writer"}, ownerU.Access)["id"].(string)
	for _, r := range []string{plain, writer} {
		grant(r, "forums", "View", true)
	}
	grant(writer, "forums", "Edit own messages", true)
	grant(writer, "forums", "Delete own messages", true)

	o := e.Owner(ownerU)
	a := e.Join(ownerU, slug, "Ann", plain)
	b := e.Join(ownerU, slug, "Bea", writer)
	for _, m := range []*testkit.Member{o, a, b} {
		m.Register(e)
	}
	key, _ := e2ee.NewConversationKey()
	conv := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "forum", "name": "Team", "members": []string{a.Person, b.Person}, "keys": o.WrapFor(e, key, 1, o.Person, a.Person, b.Person)}, ownerU.Access)["id"].(string)

	annMsg := a.Send(e, org, conv, 1, key, "from Ann", 201)["id"].(string)
	beaMsg := b.Send(e, org, conv, 1, key, "from Bea", 201)["id"].(string)

	edit := func(m *testkit.Member, of, text string, want int) {
		t.Helper()
		id := uuid.Must(uuid.NewV7()).String()
		e.MustDo(want, "POST", "/conversations/"+conv+"/messages", map[string]any{"id": id, "type": "edit", "reply_to": of, "epoch": 1, "device": m.DevID, "envelope": m.Seal(org, conv, id, 1, key, text)}, m.Access)
	}
	// without the own rights, Ann can neither edit nor delete her own message
	edit(a, annMsg, "Ann again", 403)
	if code, _ := e.Do("DELETE", "/conversations/"+conv+"/messages/"+annMsg, nil, a.Access); code != 403 {
		t.Fatalf("a role without the own right deleted its own message: %d", code)
	}
	// Bea's role has them: she edits and deletes her own...
	edit(b, beaMsg, "Bea edited", 201)
	// ...but not Ann's
	edit(b, annMsg, "hijack", 403)
	if code, _ := e.Do("DELETE", "/conversations/"+conv+"/messages/"+annMsg, nil, b.Access); code != 403 {
		t.Fatalf("someone deleted another person's message with only the own right: %d", code)
	}
	// an edit needs an original
	edit(b, uuid.NewString(), "ghost", 404)
	if code, _ := e.Do("DELETE", "/conversations/"+conv+"/messages/"+beaMsg, nil, b.Access); code != 204 {
		t.Fatalf("own message not deleted: %d", code)
	}
	// give Ann the right and she can
	grant(plain, "forums", "Delete own messages", true)
	if code, _ := e.Do("DELETE", "/conversations/"+conv+"/messages/"+annMsg, nil, a.Access); code != 204 {
		t.Fatalf("own message not deleted after the right was given: %d", code)
	}
}

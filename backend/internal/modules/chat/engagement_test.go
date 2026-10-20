package chat_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/e2ee"
	"hub/internal/testkit"
)

// One vote per person per poll, at most one of each reaction, locks only by the poll's author or a moderator, and a like
// or vote changed outside Hub is hidden and raises the "Tampered records" alert.
func TestLikesAndVotesAreOnePerPersonAndTamperProof(t *testing.T) {
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
	send := func(m *testkit.Member, typ, text string, replyTo string) (int, string) {
		id := uuid.Must(uuid.NewV7()).String()
		body := map[string]any{"id": id, "epoch": 1, "device": m.DevID, "envelope": m.Seal(org, conv, id, 1, key, text), "type": typ}
		if replyTo != "" {
			body["reply_to"] = replyTo
		}
		code, _ := e.Do("POST", "/conversations/"+conv+"/messages", body, m.Access)
		return code, id
	}
	_, poll := send(o, "poll", `{"question":"Lunch?","options":[{"id":"o1","label":"Yes"}]}`, "")
	_, post := send(o, "text", `{"text":"hello"}`, "")

	// one vote each; a second needs the first taken back
	if c, _ := send(a, "vote", `{"option":"o1"}`, poll); c != 201 {
		t.Fatalf("vote: %d", c)
	}
	if c, _ := send(a, "vote", `{"option":"o1"}`, poll); c != 409 {
		t.Fatalf("second vote: %d", c)
	}
	if c, _ := send(a, "vote", `{"option":"o1"}`, post); c != 400 {
		t.Fatalf("vote on a plain message: %d", c)
	}
	if c, _ := send(a, "vote", `{"option":"o1"}`, ""); c != 400 {
		t.Fatalf("vote on nothing: %d", c)
	}
	// like and love, but no third reaction
	_, like := send(a, "reaction", `{"kind":"like"}`, post)
	if c, _ := send(a, "reaction", `{"kind":"love"}`, post); c != 201 {
		t.Fatalf("love: %d", c)
	}
	if c, _ := send(a, "reaction", `{"kind":"like"}`, post); c != 409 {
		t.Fatalf("third reaction: %d", c)
	}
	// Amy did not make the poll and is not a moderator: she cannot lock it
	if c, _ := send(a, "poll_lock", `{"locked":true}`, poll); c != 403 {
		t.Fatalf("lock by a member: %d", c)
	}
	if c, _ := send(o, "poll_lock", `{"locked":true}`, poll); c != 201 {
		t.Fatalf("lock by the author: %d", c)
	}

	// someone with database access rewrites Amy's like (as if to forge or move it)
	root, err := pgx.Connect(context.Background(), "postgres://hubadmin@/hub?host=/tmp&port=5433")
	if err != nil {
		t.Skip("superuser not reachable: ", err)
	}
	defer root.Close(context.Background())
	_, _ = root.Exec(context.Background(), "set session_replication_role = replica")
	if _, err := root.Exec(context.Background(), `update hub.messages set created_at = created_at - interval '1 day' where id=$1`, like); err != nil {
		t.Skip("cannot simulate tampering: ", err)
	}
	// the like is hidden from everyone reading the conversation...
	for _, m := range e.MustDo(200, "GET", "/conversations/"+conv+"/messages?limit=100", nil, o.Access)["messages"].([]any) {
		if m.(map[string]any)["id"] == like {
			t.Fatal("a tampered like was served")
		}
	}
	// ...and the security logs page shows a Tampered records alert
	found := false
	for _, x := range e.MustDo(200, "GET", "/security/alerts", nil, ownerU.Access)["alerts"].([]any) {
		found = found || x.(map[string]any)["rule"] == "r8"
	}
	if !found {
		t.Fatal("no Tampered records alert")
	}
}

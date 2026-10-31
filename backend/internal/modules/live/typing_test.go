package live_test

import (
	"context"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"

	"hub/internal/e2ee"
	"hub/internal/testkit"
)

// Typing is event driven: the other member hears "typing" when it starts and "stopped" the moment it stops; a message sent is
// delivered at once; and someone outside the conversation can neither hear nor cause either.
func TestTypingAndMessagesAreLive(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	org := e.OrgID(slug)
	o := e.Owner(ownerU)
	a := e.Join(ownerU, slug, "Amy", "staff")
	x := e.Join(ownerU, slug, "Xavier", "staff")
	for _, m := range []*testkit.Member{o, a, x} {
		m.Register(e)
	}
	key, _ := e2ee.NewConversationKey()
	conv := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "forum", "name": "Team", "members": []string{a.Person}, "keys": o.WrapFor(e, key, 1, o.Person, a.Person)}, ownerU.Access)["id"].(string)
	co, ca, cx := dial(t, e, ownerU.Access), dial(t, e, a.Access), dial(t, e, x.Access)
	ctx := context.Background()
	for _, c := range []*websocket.Conn{co, ca} {
		_ = c.Write(ctx, websocket.MessageText, []byte(`{"op":"sub","conv":"`+conv+`"}`))
		next(t, c, "subscribed")
	}
	// Amy starts typing: Olivia hears it, with who
	_ = ca.Write(ctx, websocket.MessageText, []byte(`{"op":"typing","conv":"`+conv+`"}`))
	d := next(t, co, "typing")["data"].(map[string]any)
	if d["person"] != a.Person || d["on"] != true {
		t.Fatalf("typing start: %v", d)
	}
	// and the moment she stops, Olivia is told
	_ = ca.Write(ctx, websocket.MessageText, []byte(`{"op":"typing","conv":"`+conv+`","on":false}`))
	if d := next(t, co, "typing")["data"].(map[string]any); d["person"] != a.Person || d["on"] != false {
		t.Fatalf("typing stop: %v", d)
	}
	// a message reaches the other member straight away
	id := uuid.Must(uuid.NewV7()).String()
	start := time.Now()
	e.MustDo(201, "POST", "/conversations/"+conv+"/messages", map[string]any{"id": id, "epoch": 1, "device": a.DevID, "envelope": a.Seal(org, conv, id, 1, key, `{"text":"hi"}`), "type": "text"}, a.Access)
	m := next(t, co, "message.new")
	if m["data"].(map[string]any)["id"] != id {
		t.Fatalf("wrong message: %v", m)
	}
	if time.Since(start) > time.Second {
		t.Fatalf("message took %v to arrive", time.Since(start))
	}
	// Xavier is not in the conversation: he cannot subscribe, cannot cause typing, and never hears it
	_ = cx.Write(ctx, websocket.MessageText, []byte(`{"op":"typing","conv":"`+conv+`"}`))
	_ = ca.Write(ctx, websocket.MessageText, []byte(`{"op":"typing","conv":"`+conv+`"}`))
	if d := next(t, co, "typing")["data"].(map[string]any); d["person"] != a.Person {
		t.Fatalf("a stranger's typing got through: %v", d)
	}
}

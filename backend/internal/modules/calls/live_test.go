package calls_test

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/coder/websocket"

	"hub/internal/callsvc"
	"hub/internal/e2ee"
	"hub/internal/modules/calls"
	"hub/internal/testkit"
)

// A call in progress shows to the people it rings and to members of its conversation, and to nobody else, so a reload never hides it.
func TestLiveCallsShowToTheRightPeople(t *testing.T) {
	e := testkit.New(t)
	svc := httptest.NewServer(callsvc.New(e.A.Cfg.CallsSecret, e.A.Rdb, nil).Handler())
	defer svc.Close()
	e.A.Cfg.CallsURLs = []string{svc.URL}
	ctx, stop := context.WithCancel(context.Background())
	defer stop()
	go calls.New(e.A).Consume(ctx)

	ownerU, slug := e.NewOrg("Olivia Owner")
	o := e.Owner(ownerU)
	a := e.Join(ownerU, slug, "Amy", "staff")
	b := e.Join(ownerU, slug, "Ben", "staff")
	for _, m := range []*testkit.Member{o, a, b} {
		m.Register(e)
	}
	key, _ := e2ee.NewConversationKey()
	dm := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "dm", "members": []string{a.Person}, "keys": o.WrapFor(e, key, 1, o.Person, a.Person)}, ownerU.Access)["id"].(string)
	// Amy is signed in (a live connection), so the call rings her
	tk := e.MustDo(200, "POST", "/realtime/ticket", nil, a.Access)["ticket"].(string)
	ws, _, err := websocket.Dial(context.Background(), "ws"+strings.TrimPrefix(e.Srv.URL, "http")+"/v1/realtime?ticket="+tk, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ws.CloseNow()
	started := e.MustDo(201, "POST", "/calls", map[string]any{"conversation": dm, "kind": "voice", "context": "dm"}, ownerU.Access)
	callID := started["call"].(map[string]any)["id"].(string)

	find := func(access string) map[string]any {
		for _, c := range e.MustDo(200, "GET", "/calls/live", nil, access)["calls"].([]any) {
			if m := c.(map[string]any); m["id"] == callID {
				return m
			}
		}
		return nil
	}
	amy := find(a.Access)
	if amy == nil || amy["ringing_me"] != true || amy["invited"] != true || amy["in_call"] != false {
		t.Fatalf("Amy should see the call ringing her: %v", amy)
	}
	host := find(ownerU.Access)
	if host == nil || host["ringing_me"] != false {
		t.Fatalf("the caller sees the call but it is not ringing them: %v", host)
	}
	if find(b.Access) != nil {
		t.Fatal("Ben, outside the conversation, must not see the call")
	}
	// once Amy declines, it no longer rings her
	e.MustDo(204, "POST", "/calls/"+callID+"/decline", nil, a.Access)
	if amy := find(a.Access); amy != nil && amy["ringing_me"] == true {
		t.Fatalf("a declined call still rings: %v", amy)
	}
}

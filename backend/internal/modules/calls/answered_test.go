package calls_test

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"

	"hub/internal/callsvc"
	"hub/internal/e2ee"
	"hub/internal/testkit"
)

// The caller stops ringing the moment the other person picks up: from the pick-up request itself, with no wait for their microphone,
// the media connection or the worker. Nothing else is running here (no call room, no worker) to prove that.
func TestCallerIsToldAtOnceThatTheCallWasPickedUp(t *testing.T) {
	e := testkit.New(t)
	svc := httptest.NewServer(callsvc.New(e.A.Cfg.CallsSecret, e.A.Rdb, nil).Handler())
	defer svc.Close()
	e.A.Cfg.CallsURLs = []string{svc.URL}

	ownerU, slug := e.NewOrg("Olivia Owner")
	o := e.Owner(ownerU)
	a := e.Join(ownerU, slug, "Amy", "staff")
	for _, m := range []*testkit.Member{o, a} {
		m.Register(e)
	}
	key, _ := e2ee.NewConversationKey()
	dm := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "dm", "members": []string{a.Person}, "keys": o.WrapFor(e, key, 1, o.Person, a.Person)}, ownerU.Access)["id"].(string)
	listen := func(access string) *websocket.Conn {
		tk := e.MustDo(200, "POST", "/realtime/ticket", nil, access)["ticket"].(string)
		ws, _, err := websocket.Dial(context.Background(), "ws"+strings.TrimPrefix(e.Srv.URL, "http")+"/v1/realtime?ticket="+tk, nil)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = ws.CloseNow() })
		return ws
	}
	callerWS, calleeWS := listen(ownerU.Access), listen(a.Access)

	started := e.MustDo(201, "POST", "/calls", map[string]any{"conversation": dm, "kind": "video", "context": "dm"}, ownerU.Access)
	callID := started["call"].(map[string]any)["id"].(string)
	read(t, calleeWS, "call.incoming")

	t0 := time.Now()
	e.MustDo(200, "POST", "/calls/"+callID+"/join", nil, a.Access)
	got := read(t, callerWS, "call.answered")
	if got["data"].(map[string]any)["call"] != callID {
		t.Fatalf("answered notice for another call: %v", got)
	}
	if d := time.Since(t0); d > 2*time.Second {
		t.Fatalf("the caller heard about the pick-up after %v", d)
	}
}

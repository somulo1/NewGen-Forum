package calls_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"

	"hub/internal/callsvc"
	"hub/internal/e2ee"
	"hub/internal/modules/calls"
	"hub/internal/testkit"
)

func connect(t *testing.T, entry map[string]any) *websocket.Conn {
	t.Helper()
	url := entry["url"].(string) + "?token=" + entry["token"].(string)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.CloseNow() })
	return c
}

func read(t *testing.T, c *websocket.Conn, typ string) map[string]any {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for {
		_, b, err := c.Read(ctx)
		if err != nil {
			t.Fatalf("waiting for %s: %v", typ, err)
		}
		var m map[string]any
		_ = json.Unmarshal(b, &m)
		if m["type"] == typ {
			return m
		}
	}
}

func send(t *testing.T, c *websocket.Conn, v any) {
	t.Helper()
	b, _ := json.Marshal(v)
	if err := c.Write(context.Background(), websocket.MessageText, b); err != nil {
		t.Fatal(err)
	}
}

func TestCallLifecycleAcrossTheSeparateService(t *testing.T) {
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

	// Amy listens for the ring
	tk := e.MustDo(200, "POST", "/realtime/ticket", nil, a.Access)["ticket"].(string)
	ws, _, err := websocket.Dial(context.Background(), "ws"+strings.TrimPrefix(e.Srv.URL, "http")+"/v1/realtime?ticket="+tk, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ws.CloseNow()

	// Ben (not in the conversation) cannot start a call into it
	if code, _ := e.Do("POST", "/calls", map[string]any{"conversation": dm, "kind": "voice", "context": "dm"}, b.Access); code != 404 {
		t.Fatalf("outsider rang a conversation: %d", code)
	}
	started := e.MustDo(201, "POST", "/calls", map[string]any{"conversation": dm, "kind": "video", "context": "dm"}, ownerU.Access)
	ring := read(t, ws, "call.incoming")
	callID := started["call"].(map[string]any)["id"].(string)
	if ring["data"].(map[string]any)["call"] != callID {
		t.Fatalf("ring for the wrong call: %v", ring)
	}
	// Ben cannot join; Amy can
	if code, _ := e.Do("POST", "/calls/"+callID+"/join", nil, b.Access); code != 404 {
		t.Fatalf("outsider joined: %d", code)
	}
	joined := e.MustDo(200, "POST", "/calls/"+callID+"/join", nil, a.Access)

	hostWS := connect(t, started)
	welcome := read(t, hostWS, "welcome")
	if welcome["you"].(map[string]any)["role"] != "host" {
		t.Fatalf("host role: %v", welcome)
	}
	amyWS := connect(t, joined)
	aw := read(t, amyWS, "welcome")
	if len(aw["peers"].([]any)) != 1 {
		t.Fatalf("Amy should see the host: %v", aw)
	}
	// a pass works once
	if _, resp, err := websocket.Dial(context.Background(), joined["url"].(string)+"?token="+joined["token"].(string), nil); err == nil || resp == nil || resp.StatusCode != 401 {
		t.Fatalf("a call pass was used twice: %v", err)
	}
	// signaling is relayed untouched and only to the addressee
	hostID := aw["peers"].([]any)[0].(map[string]any)["id"].(string)
	send(t, amyWS, map[string]any{"type": "offer", "to": hostID, "data": map[string]string{"sdp": "v=0..."}})
	off := read(t, hostWS, "offer")
	if off["from"] != aw["you"].(map[string]any)["id"] {
		t.Fatalf("offer: %v", off)
	}
	// only the host can moderate
	send(t, amyWS, map[string]any{"type": "mod", "action": "mute", "target": hostID})
	read(t, amyWS, "error")
	amyID := aw["you"].(map[string]any)["id"].(string)
	send(t, hostWS, map[string]any{"type": "mod", "action": "mute", "target": amyID})
	if m := read(t, amyWS, "mod"); m["action"] != "mute" {
		t.Fatalf("mod: %v", m)
	}

	// the history follows the real events
	waitStatus := func(want string) map[string]any {
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			l := e.MustDo(200, "GET", "/calls/history", nil, ownerU.Access)["calls"].([]any)
			if len(l) == 1 && l[0].(map[string]any)["status"] == want {
				return l[0].(map[string]any)
			}
			time.Sleep(100 * time.Millisecond)
		}
		t.Fatalf("history never reached %q", want)
		return nil
	}
	waitStatus("active")
	send(t, hostWS, map[string]any{"type": "mod", "action": "end"})
	read(t, amyWS, "ended")
	end := waitStatus("ended")
	if end["direction"] != "out" {
		t.Fatalf("direction: %v", end)
	}
	if in := e.MustDo(200, "GET", "/calls/history", nil, a.Access)["calls"].([]any); len(in) != 1 || in[0].(map[string]any)["direction"] != "in" {
		t.Fatalf("Amy's history: %v", in)
	}
}

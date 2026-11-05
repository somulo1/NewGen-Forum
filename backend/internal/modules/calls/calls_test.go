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

// Hanging up before anyone answers stops every phone that is ringing, at once; a member who leaves does not end it for the rest.
func TestHangingUpBeforeAnswerStopsTheRingAtOnce(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	o := e.Owner(ownerU)
	a := e.Join(ownerU, slug, "Amy", "staff")
	b := e.Join(ownerU, slug, "Ben", "staff")
	for _, m := range []*testkit.Member{o, a, b} {
		m.Register(e)
	}
	listen := func(m *testkit.Member) *websocket.Conn {
		tk := e.MustDo(200, "POST", "/realtime/ticket", nil, m.Access)["ticket"].(string)
		ws, _, err := websocket.Dial(context.Background(), "ws"+strings.TrimPrefix(e.Srv.URL, "http")+"/v1/realtime?ticket="+tk, nil)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = ws.CloseNow() })
		return ws
	}
	wsA, wsB := listen(a), listen(b)
	key, _ := e2ee.NewConversationKey()

	// a private call: the caller hangs up while it rings
	dm := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "dm", "members": []string{a.Person}, "keys": o.WrapFor(e, key, 1, o.Person, a.Person)}, ownerU.Access)["id"].(string)
	call := e.MustDo(201, "POST", "/calls", map[string]any{"conversation": dm, "kind": "voice", "context": "dm"}, ownerU.Access)["call"].(map[string]any)["id"].(string)
	read(t, wsA, "call.incoming")
	t0 := time.Now()
	if r := e.MustDo(200, "POST", "/calls/"+call+"/leave", nil, ownerU.Access); r["cancelled"] != true {
		t.Fatalf("leave: %v", r)
	}
	read(t, wsA, "call.ended")
	if d := time.Since(t0); d > 2*time.Second {
		t.Fatalf("the ring took %v to stop", d)
	}
	if code, _ := e.Do("POST", "/calls/"+call+"/leave", nil, b.Access); code != 404 {
		t.Fatalf("a stranger left a call: %d", code)
	}

	// deleting a call from my history is for good, and only mine: Amy still has hers
	has := func(m *testkit.Member) bool {
		for _, c := range e.MustDo(200, "GET", "/calls/history", nil, m.Access)["calls"].([]any) {
			if c.(map[string]any)["id"] == call {
				return true
			}
		}
		return false
	}
	if !has(o) || !has(a) {
		t.Fatalf("the finished call is missing from a history: %v %v", has(o), has(a))
	}
	if r := e.MustDo(200, "POST", "/calls/history/delete", map[string]any{"ids": []string{call}}, ownerU.Access); r["deleted"].(float64) != 1 {
		t.Fatalf("delete: %v", r)
	}
	if has(o) || !has(a) {
		t.Fatalf("after deleting: mine %v, Amy's %v", has(o), has(a))
	}
	if r := e.MustDo(200, "POST", "/calls/history/delete", map[string]any{"ids": []string{call}}, b.Access); r["deleted"].(float64) != 0 {
		t.Fatalf("a stranger deleted from a history: %v", r)
	}

	// a forum call: members are ringing; one of them leaving does not end it, the caller leaving before anyone answers does
	fk, _ := e2ee.NewConversationKey()
	forum := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "forum", "name": "Team", "members": []string{a.Person, b.Person}, "keys": o.WrapFor(e, fk, 1, o.Person, a.Person, b.Person)}, ownerU.Access)["id"].(string)
	fc := e.MustDo(201, "POST", "/calls", map[string]any{"conversation": forum, "kind": "voice", "context": "forum"}, ownerU.Access)["call"].(map[string]any)["id"].(string)
	read(t, wsA, "call.started")
	read(t, wsB, "call.started")
	if r := e.MustDo(200, "POST", "/calls/"+fc+"/leave", nil, a.Access); r["cancelled"] != false {
		t.Fatalf("a member leaving ended the call: %v", r)
	}
	t0 = time.Now()
	if r := e.MustDo(200, "POST", "/calls/"+fc+"/leave", nil, ownerU.Access); r["cancelled"] != true {
		t.Fatalf("the caller leaving: %v", r)
	}
	read(t, wsB, "call.ended")
	if d := time.Since(t0); d > 2*time.Second {
		t.Fatalf("the forum ring took %v to stop", d)
	}
}

// When one person hangs up and only one is left in a private call, the call ends for the last one too, at once.
func TestTheLastPersonLeftInACallIsHungUp(t *testing.T) {
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
	for _, m := range []*testkit.Member{o, a} {
		m.Register(e)
	}
	listen := func(m *testkit.Member) *websocket.Conn {
		tk := e.MustDo(200, "POST", "/realtime/ticket", nil, m.Access)["ticket"].(string)
		ws, _, err := websocket.Dial(context.Background(), "ws"+strings.TrimPrefix(e.Srv.URL, "http")+"/v1/realtime?ticket="+tk, nil)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = ws.CloseNow() })
		return ws
	}
	wsO, wsA := listen(o), listen(a)
	key, _ := e2ee.NewConversationKey()
	dm := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "dm", "members": []string{a.Person}, "keys": o.WrapFor(e, key, 1, o.Person, a.Person)}, ownerU.Access)["id"].(string)
	started := e.MustDo(201, "POST", "/calls", map[string]any{"conversation": dm, "kind": "voice", "context": "dm"}, ownerU.Access)
	read(t, wsA, "call.incoming")
	callID := started["call"].(map[string]any)["id"].(string)
	joined := e.MustDo(200, "POST", "/calls/"+callID+"/join", nil, a.Access)
	hostRoom, amyRoom := connect(t, started), connect(t, joined)
	read(t, hostRoom, "welcome")
	read(t, amyRoom, "welcome")
	read(t, wsO, "call.answered") // the instant notice, when Amy picked up
	read(t, wsO, "call.answered") // and the one that follows once the room has recorded both people

	// Amy hangs up: the call is told to Olivia straight away, so her side hangs up too
	t0 := time.Now()
	e.MustDo(200, "POST", "/calls/"+callID+"/leave", nil, a.Access)
	read(t, wsO, "call.ended")
	if d := time.Since(t0); d > 2*time.Second {
		t.Fatalf("the last person was hung up after %v", d)
	}
	h := e.MustDo(200, "GET", "/calls/history", nil, ownerU.Access)["calls"].([]any)
	if len(h) == 0 || h[0].(map[string]any)["status"] != "ended" {
		t.Fatalf("history: %v", h)
	}
}

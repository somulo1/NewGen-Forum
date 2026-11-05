package calls_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"

	"hub/internal/callsvc"
	"hub/internal/e2ee"
	"hub/internal/testkit"
)

// online opens the realtime connection that makes a person "online", which a private call requires of the person being rung.
func online(t *testing.T, e *testkit.Env, access string) {
	t.Helper()
	tk := e.MustDo(200, "POST", "/realtime/ticket", nil, access)["ticket"].(string)
	ws, _, err := websocket.Dial(context.Background(), "ws"+strings.TrimPrefix(e.Srv.URL, "http")+"/v1/realtime?ticket="+tk, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ws.CloseNow() })
}

// tap reads a connection in the background so a test can wait for a message with a time limit without closing the connection (the
// websocket library closes a connection whose read is cancelled).
type tap struct{ ch chan map[string]any }

func watch(c *websocket.Conn) *tap {
	tp := &tap{ch: make(chan map[string]any, 256)}
	go func() {
		for {
			_, b, err := c.Read(context.Background())
			if err != nil {
				return
			}
			var m map[string]any
			if json.Unmarshal(b, &m) == nil {
				tp.ch <- m
			}
		}
	}()
	return tp
}

// within waits up to d for a message of the given type; false when none comes.
func (tp *tap) within(typ string, d time.Duration) (map[string]any, bool) {
	t := time.NewTimer(d)
	defer t.Stop()
	for {
		select {
		case m := <-tp.ch:
			if m["type"] == typ {
				return m, true
			}
		case <-t.C:
			return nil, false
		}
	}
}

// A device whose connection drops keeps its place for a while: nobody is told it left, it comes back with its secret, and what was sent
// to it meanwhile reaches it. One that hangs up properly is gone at once, and one that never returns is removed when the time is up.
func TestADroppedDeviceKeepsItsPlaceAndComesBack(t *testing.T) {
	e := testkit.New(t)
	svc := callsvc.New(e.A.Cfg.CallsSecret, e.A.Rdb, nil)
	svc.Grace = 1500 * time.Millisecond
	hs := httptest.NewServer(svc.Handler())
	defer hs.Close()
	e.A.Cfg.CallsURLs = []string{hs.URL}

	ownerU, slug := e.NewOrg("Olivia Owner")
	o := e.Owner(ownerU)
	a := e.Join(ownerU, slug, "Amy", "staff")
	for _, m := range []*testkit.Member{o, a} {
		m.Register(e)
	}
	key, _ := e2ee.NewConversationKey()
	dm := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "dm", "members": []string{a.Person}, "keys": o.WrapFor(e, key, 1, o.Person, a.Person)}, ownerU.Access)["id"].(string)
	online(t, e, a.Access)
	started := e.MustDo(201, "POST", "/calls", map[string]any{"conversation": dm, "kind": "voice", "context": "dm"}, ownerU.Access)
	callID := started["call"].(map[string]any)["id"].(string)
	joined := e.MustDo(200, "POST", "/calls/"+callID+"/join", nil, a.Access)

	hostWS := connect(t, started)
	read(t, hostWS, "welcome")
	host := watch(hostWS)
	amyWS := connect(t, joined)
	aw := read(t, amyWS, "welcome")
	amyID := aw["you"].(map[string]any)["id"].(string)
	secret, _ := aw["resume"].(string)
	if secret == "" {
		t.Fatal("the welcome carries no secret to resume with")
	}
	room := started["url"].(string)
	resumeURL := func(k string) string {
		q := url.Values{"resume": {amyID}, "room": {callID}, "key": {k}}
		return room + "?" + q.Encode()
	}
	dial := func(u string) (*websocket.Conn, int, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		c, resp, err := websocket.Dial(ctx, u, nil)
		code := 0
		if resp != nil {
			code = resp.StatusCode
		}
		return c, code, err
	}

	// the wrong secret gets nowhere
	if _, code, err := dial(resumeURL("not-the-secret")); err == nil || code != 410 {
		t.Fatalf("resumed with the wrong secret: %v %d", err, code)
	}

	// Amy's connection dies without a goodbye: the host is NOT told she left
	_ = amyWS.CloseNow()
	if _, left := host.within("peer-left", 500*time.Millisecond); left {
		t.Fatal("the host was told Amy left the moment her connection dropped")
	}
	// the host sends her an offer while she is away; she comes back with her secret
	send(t, hostWS, map[string]any{"type": "offer", "to": amyID, "data": map[string]string{"sdp": "v=0..."}})
	back, _, err := dial(resumeURL(secret))
	if err != nil {
		t.Fatalf("could not come back: %v", err)
	}
	t.Cleanup(func() { _ = back.CloseNow() })
	amy := watch(back)
	r, _ := amy.within("resumed", 3*time.Second)
	if r == nil {
		t.Fatal("no resumed message")
	}
	if r["you"].(map[string]any)["id"] != amyID || len(r["peers"].([]any)) != 1 {
		t.Fatalf("resumed: %v", r)
	}
	if off, ok := amy.within("offer", 2*time.Second); !ok || off["data"] == nil {
		t.Fatal("what the host sent while she was away did not reach her")
	}
	// nobody was told anything, and the room still works both ways
	if _, left := host.within("peer-left", 300*time.Millisecond); left {
		t.Fatal("the host was told Amy left although she came back")
	}
	send(t, back, map[string]any{"type": "state", "mic": false})
	if st, ok := host.within("state", 2*time.Second); !ok || st["peer"].(map[string]any)["mic"] != false {
		t.Fatal("the host does not hear Amy's changes after she came back")
	}

	// hanging up properly removes her at once
	t0 := time.Now()
	_ = back.Close(websocket.StatusNormalClosure, "bye")
	if _, ok := host.within("peer-left", 2*time.Second); !ok || time.Since(t0) > 1200*time.Millisecond {
		t.Fatalf("hanging up did not remove her at once (%v)", time.Since(t0))
	}
	// and a place that was given up cannot be taken back
	if _, code, err := dial(resumeURL(secret)); err == nil || code != 410 {
		t.Fatalf("resumed a place that no longer exists: %v %d", err, code)
	}
}

// Someone who never comes back is removed when the grace period ends, and the others are told then.
func TestADeviceThatNeverReturnsIsRemovedAfterTheGrace(t *testing.T) {
	e := testkit.New(t)
	svc := callsvc.New(e.A.Cfg.CallsSecret, e.A.Rdb, nil)
	svc.Grace = 600 * time.Millisecond
	hs := httptest.NewServer(svc.Handler())
	defer hs.Close()
	e.A.Cfg.CallsURLs = []string{hs.URL}
	ownerU, slug := e.NewOrg("Olivia Owner")
	o := e.Owner(ownerU)
	a := e.Join(ownerU, slug, "Amy", "staff")
	for _, m := range []*testkit.Member{o, a} {
		m.Register(e)
	}
	key, _ := e2ee.NewConversationKey()
	dm := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "dm", "members": []string{a.Person}, "keys": o.WrapFor(e, key, 1, o.Person, a.Person)}, ownerU.Access)["id"].(string)
	online(t, e, a.Access)
	started := e.MustDo(201, "POST", "/calls", map[string]any{"conversation": dm, "kind": "voice", "context": "dm"}, ownerU.Access)
	callID := started["call"].(map[string]any)["id"].(string)
	joined := e.MustDo(200, "POST", "/calls/"+callID+"/join", nil, a.Access)
	hostWS := connect(t, started)
	read(t, hostWS, "welcome")
	host := watch(hostWS)
	amyWS := connect(t, joined)
	read(t, amyWS, "welcome")

	t0 := time.Now()
	_ = amyWS.CloseNow()
	if _, ok := host.within("peer-left", 3*time.Second); !ok {
		t.Fatal("a device that never came back was not removed")
	}
	if d := time.Since(t0); d < 500*time.Millisecond {
		t.Fatalf("removed after only %v: the grace period was skipped", d)
	}
}

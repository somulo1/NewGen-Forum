package live_test

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"

	"hub/internal/testkit"
)

var tokenRe = regexp.MustCompile(`/join\?t=(\S+)`)

func dial(t *testing.T, e *testkit.Env, access string) *websocket.Conn {
	t.Helper()
	tk := e.MustDo(200, "POST", "/realtime/ticket", nil, access)["ticket"].(string)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(e.Srv.URL, "http")+"/v1/realtime?ticket="+tk, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.CloseNow() })
	return c
}

func next(t *testing.T, c *websocket.Conn, want string) map[string]any {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for {
		_, b, err := c.Read(ctx)
		if err != nil {
			t.Fatalf("waiting for %s: %v", want, err)
		}
		var m map[string]any
		_ = json.Unmarshal(b, &m)
		if m["type"] == want {
			return m
		}
	}
}

func TestRealtimeDeliversAndGuards(t *testing.T) {
	e := testkit.New(t)
	owner, slug := e.NewOrg("Olivia Owner")
	// a second member
	email := testkit.Uniq()
	out := e.MustDo(201, "POST", "/people", map[string]any{"people": []any{map[string]any{"first_name": "Sam", "last_name": "T", "email": email, "phone": "+254700000001", "date_of_birth": "1990-04-05", "gender": "Male"}}}, owner.Access)
	sid := out["results"].([]any)[0].(map[string]any)["id"].(string)
	tok := tokenRe.FindStringSubmatch(e.Dev.To(email)[0].Body)[1]
	e.Signup("Sam", email)
	e.MustDo(200, "POST", "/invites/accept", map[string]any{"token": tok}, e.Login(email, "").Access)
	sam := e.Login(email, slug)

	// tickets are single use
	tk := e.MustDo(200, "POST", "/realtime/ticket", nil, owner.Access)["ticket"].(string)
	url := "ws" + strings.TrimPrefix(e.Srv.URL, "http") + "/v1/realtime?ticket=" + tk
	ctx := context.Background()
	c1, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c1.CloseNow()
	if _, _, err := websocket.Dial(ctx, url, nil); err == nil {
		t.Fatal("a ticket worked twice")
	}
	if _, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(e.Srv.URL, "http")+"/v1/realtime?ticket=forged", nil); err == nil {
		t.Fatal("forged ticket accepted")
	}

	// owner's socket sees changes made to the org in real time
	co := dial(t, e, owner.Access)
	cs := dial(t, e, sam.Access)
	next(t, co, "presence")
	e.MustDo(204, "POST", "/people/"+sid+"/status", map[string]any{"action": "deactivate"}, owner.Access)
	ev := next(t, co, "person.status")
	if ev["org"] == "" || ev["data"].(map[string]any)["status"] != "inactive" {
		t.Fatalf("bad event %v", ev)
	}
	// and the deactivated person's socket is closed by the server
	next(t, cs, "session.ended")
	ctx2, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	for {
		if _, _, err := cs.Read(ctx2); err != nil {
			break
		}
	}
	if ctx2.Err() != nil {
		t.Fatal("socket stayed open after the person was deactivated")
	}

	// subscribing to a conversation you are not in is refused
	_ = co.Write(ctx, websocket.MessageText, []byte(`{"op":"sub","conv":"01a154f7-1480-75de-9077-b2acb71da8d7"}`))
	if m := next(t, co, "error"); m["code"] != "forbidden" {
		t.Fatalf("expected forbidden: %v", m)
	}
}

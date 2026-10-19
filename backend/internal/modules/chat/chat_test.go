package chat_test

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/e2ee"
	"hub/internal/modules/keys"
	"hub/internal/testkit"
)

var tokenRe = regexp.MustCompile(`/join\?t=(\S+)`)

type member struct {
	*testkit.User
	person string
	dev    *e2ee.Device
	devID  string
	acct   string
}

func (m *member) pub() map[string]any {
	return map[string]any{"label": "test", "identity_key": b64(m.dev.Public().Identity), "signing_key": b64(m.dev.Public().Signing)}
}

func b64(b []byte) string { return base64.StdEncoding.EncodeToString(b) }

func newMember(t *testing.T, e *testkit.Env, owner *testkit.User, slug, first string) *member {
	t.Helper()
	email := testkit.Uniq()
	out := e.MustDo(201, "POST", "/people", map[string]any{"people": []any{map[string]any{"first_name": first, "last_name": "T", "email": email, "phone": "+254700000001", "date_of_birth": "1990-04-05", "gender": "Male"}}}, owner.Access)
	pid := out["results"].([]any)[0].(map[string]any)["id"].(string)
	tok := tokenRe.FindStringSubmatch(e.Dev.To(email)[0].Body)[1]
	e.Signup(first, email)
	e.MustDo(200, "POST", "/invites/accept", map[string]any{"token": tok}, e.Login(email, "").Access)
	u := e.Login(email, slug)
	return &member{User: u, person: pid}
}

func (m *member) register(t *testing.T, e *testkit.Env) {
	t.Helper()
	d, err := e2ee.NewDevice()
	if err != nil {
		t.Fatal(err)
	}
	m.dev = d
	me := e.MustDo(200, "GET", "/me", nil, m.Access)
	m.acct = me["account"].(string)
	acct, _ := uuid.Parse(m.acct)
	pub := d.Public()
	bind := ed25519.Sign(d.Sign, keys.BindingMessage(pub.Identity, acct))
	body := m.pub()
	body["binding_sig"] = b64(bind)
	// a binding made for another account must be refused
	bad := ed25519.Sign(d.Sign, keys.BindingMessage(pub.Identity, uuid.New()))
	if code, _ := e.Do("POST", "/e2ee/devices", map[string]any{"label": "x", "identity_key": body["identity_key"], "signing_key": body["signing_key"], "binding_sig": b64(bad)}, m.Access); code != 400 {
		t.Fatalf("forged binding accepted: %d", code)
	}
	m.devID = e.MustDo(201, "POST", "/e2ee/devices", body, m.Access)["id"].(string)
}

type wrapped struct {
	Epoch  int    `json:"epoch"`
	Device string `json:"device"`
	Key    string `json:"wrapped_key"`
}

// wrapFor wraps a conversation key for every active device of the given people.
func wrapFor(t *testing.T, e *testkit.Env, by *member, convKey []byte, epoch int, conv string, people ...string) []map[string]any {
	t.Helper()
	var out []map[string]any
	for _, pid := range people {
		devs := e.MustDo(200, "GET", "/e2ee/people/"+pid+"/devices", nil, by.Access)["devices"].([]any)
		for _, d := range devs {
			dm := d.(map[string]any)
			ik, _ := base64.StdEncoding.DecodeString(dm["identity_key"].(string))
			w, err := e2ee.Wrap(convKey, ik, []byte("ctx"))
			if err != nil {
				t.Fatal(err)
			}
			out = append(out, map[string]any{"epoch": epoch, "device": dm["id"], "wrapped_key": b64(w)})
		}
	}
	return out
}

func (m *member) send(t *testing.T, e *testkit.Env, org, conv string, epoch int, key []byte, text string, want int) map[string]any {
	t.Helper()
	id := uuid.Must(uuid.NewV7())
	env, err := m.dev.Seal(key, []byte(text), e2ee.Context{Org: org, Conversation: conv, Sender: m.acct, MessageID: id.String(), Epoch: epoch})
	if err != nil {
		t.Fatal(err)
	}
	code, out := e.Do("POST", "/conversations/"+conv+"/messages", map[string]any{"id": id, "epoch": epoch, "device": m.devID, "envelope": b64(env)}, m.Access)
	if code != want {
		t.Fatalf("send %q: want %d got %d %v", text, want, code, out)
	}
	return out
}

func orgID(t *testing.T, e *testkit.Env, slug string) string {
	var id string
	_ = e.A.DB.Tx(context.Background(), func(tx pgx.Tx) error {
		return tx.QueryRow(context.Background(), `select id::text from hub.orgs where slug=$1`, slug).Scan(&id)
	})
	return id
}

func (m *member) readKey(t *testing.T, e *testkit.Env, conv string, epoch int) []byte {
	t.Helper()
	ks := e.MustDo(200, "GET", "/conversations/"+conv+"/keys", nil, m.Access)["keys"].([]any)
	for _, k := range ks {
		km := k.(map[string]any)
		if int(km["epoch"].(float64)) == epoch && km["device"] == m.devID {
			w, _ := base64.StdEncoding.DecodeString(km["wrapped_key"].(string))
			key, err := m.dev.Unwrap(w, []byte("ctx"))
			if err != nil {
				t.Fatal(err)
			}
			return key
		}
	}
	return nil
}

func TestEncryptedForumLifecycle(t *testing.T) {
	e := testkit.New(t)
	owner, slug := e.NewOrg("Olivia Owner")
	org := orgID(t, e, slug)
	o := &member{User: owner}
	me := e.MustDo(200, "GET", "/me", nil, owner.Access)
	o.person = me["person"].(string)
	a, b, c := newMember(t, e, owner, slug, "Amy"), newMember(t, e, owner, slug, "Ben"), newMember(t, e, owner, slug, "Cy")
	for _, m := range []*member{o, a, b, c} {
		m.register(t, e)
	}

	key1, _ := e2ee.NewConversationKey()
	create := func(keys []map[string]any) (int, map[string]any) {
		return e.Do("POST", "/conversations", map[string]any{"kind": "forum", "name": "Risk & Compliance", "about": "secret stuff", "members": []string{a.person, b.person}, "keys": keys}, owner.Access)
	}
	// a key set that leaves out one member's device is refused
	if code, _ := create(wrapFor(t, e, o, key1, 1, "", o.person, a.person)); code != 409 {
		t.Fatalf("incomplete keys accepted: %d", code)
	}
	// ... and so is one that includes an outsider's device
	if code, _ := create(wrapFor(t, e, o, key1, 1, "", o.person, a.person, b.person, c.person)); code != 400 {
		t.Fatalf("outsider device accepted: %d", code)
	}
	code, conv := create(wrapFor(t, e, o, key1, 1, "", o.person, a.person, b.person))
	if code != 201 {
		t.Fatalf("create: %d %v", code, conv)
	}
	cid := conv["id"].(string)
	if conv["name"] != "Risk & Compliance" || int(conv["member_count"].(float64)) != 3 {
		t.Fatalf("bad conversation %v", conv)
	}

	// owner sends; a member opens it on their own device; the server stores nothing readable
	secret := "the quarterly numbers are worse than we said"
	o.send(t, e, org, cid, 1, key1, secret, 201)
	ak := a.readKey(t, e, cid, 1)
	hist := e.MustDo(200, "GET", "/conversations/"+cid+"/messages", nil, a.Access)["messages"].([]any)
	if len(hist) != 1 {
		t.Fatalf("history %v", hist)
	}
	m0 := hist[0].(map[string]any)
	env, _ := base64.StdEncoding.DecodeString(m0["envelope"].(string))
	sk := owner2signing(t, e, a, o.person, m0["sender_device"].(string))
	pt, err := e2ee.Open(ak, env, sk, e2ee.Context{Org: org, Conversation: cid, Sender: o.acct, MessageID: m0["id"].(string), Epoch: 1})
	if err != nil || string(pt) != secret {
		t.Fatalf("member could not read the message: %v %q", err, pt)
	}
	var raw []byte
	_ = e.A.DB.Tx(context.Background(), func(tx pgx.Tx) error {
		oid, _ := uuid.Parse(org)
		return e.A.DB.WithOrg(context.Background(), oid, func(tx2 pgx.Tx) error {
			return tx2.QueryRow(context.Background(), `select ciphertext from hub.messages where id=$1`, m0["id"]).Scan(&raw)
		})
	})
	if len(raw) == 0 || bytes.Contains(raw, []byte("quarterly")) {
		t.Fatal("the server holds readable message text")
	}
	// the forum's name is not readable in the database either
	_ = e.A.DB.Tx(context.Background(), func(tx pgx.Tx) error {
		oid, _ := uuid.Parse(org)
		return e.A.DB.WithOrg(context.Background(), oid, func(tx2 pgx.Tx) error {
			var ne []byte
			if err := tx2.QueryRow(context.Background(), `select name_enc from hub.conversations where id=$1`, cid).Scan(&ne); err != nil {
				t.Fatal(err)
			}
			if bytes.Contains(ne, []byte("Risk")) {
				t.Fatal("forum name stored in clear")
			}
			return nil
		})
	})
	// an outsider (not a member) sees nothing
	if code, _ := e.Do("GET", "/conversations/"+cid+"/messages", nil, c.Access); code != 404 {
		t.Fatalf("non-member read history: %d", code)
	}
	if code, _ := e.Do("POST", "/conversations/"+cid+"/messages", map[string]any{"id": uuid.Must(uuid.NewV7()), "epoch": 1, "device": c.devID, "envelope": m0["envelope"]}, c.Access); code != 404 {
		t.Fatalf("non-member sent: %d", code)
	}
	// forged sender: a member cannot claim someone else's device
	if code, _ := e.Do("POST", "/conversations/"+cid+"/messages", map[string]any{"id": uuid.Must(uuid.NewV7()), "epoch": 1, "device": o.devID, "envelope": m0["envelope"]}, a.Access); code != 400 {
		t.Fatalf("sent as someone else's device: %d", code)
	}
	// wrong epoch / garbage / duplicate are refused
	a.send(t, e, org, cid, 2, ak, "x", 409)
	if code, _ := e.Do("POST", "/conversations/"+cid+"/messages", map[string]any{"id": uuid.Must(uuid.NewV7()), "epoch": 1, "device": a.devID, "envelope": b64([]byte("not an envelope at all, plainly"))}, a.Access); code != 400 {
		t.Fatalf("garbage accepted: %d", code)
	}
	// unread + read
	cl := e.MustDo(200, "GET", "/conversations?kind=forum", nil, a.Access)["conversations"].([]any)
	if len(cl) != 1 || int(cl[0].(map[string]any)["unread"].(float64)) != 1 {
		t.Fatalf("unread wrong: %v", cl)
	}
	e.MustDo(204, "POST", "/conversations/"+cid+"/read", map[string]any{"message": m0["id"]}, a.Access)
	cl = e.MustDo(200, "GET", "/conversations?kind=forum", nil, a.Access)["conversations"].([]any)
	if int(cl[0].(map[string]any)["unread"].(float64)) != 0 {
		t.Fatalf("still unread: %v", cl)
	}

	// Ben leaves the room: the key rotates, and he cannot read what is said afterwards
	key2, _ := e2ee.NewConversationKey()
	if code, _ := e.Do("DELETE", "/conversations/"+cid+"/members/"+b.person, map[string]any{"device": o.devID, "keys": wrapFor(t, e, o, key2, 2, cid, o.person)}, owner.Access); code != 409 {
		t.Fatalf("removal without keys for everyone who stays: %d", code)
	}
	e.MustDo(204, "DELETE", "/conversations/"+cid+"/members/"+b.person, map[string]any{"device": o.devID, "keys": wrapFor(t, e, o, key2, 2, cid, o.person, a.person)}, owner.Access)
	o.send(t, e, org, cid, 2, key2, "after Ben left", 201)
	if code, _ := e.Do("GET", "/conversations/"+cid+"/messages", nil, b.Access); code != 404 {
		t.Fatalf("removed member still reads: %d", code)
	}
	if code, _ := e.Do("GET", "/conversations/"+cid+"/keys", nil, b.Access); code != 404 {
		t.Fatalf("removed member still gets keys: %d", code)
	}
	if a.readKey(t, e, cid, 2) == nil || !bytes.Equal(a.readKey(t, e, cid, 2), key2) {
		t.Fatal("remaining member did not get the new key")
	}

	// Cy is added: gets the current key only, and cannot see the history from before
	e.MustDo(200, "POST", "/conversations/"+cid+"/members", map[string]any{"device": o.devID, "people": []string{c.person}, "keys": wrapFor(t, e, o, key2, 2, cid, c.person)}, owner.Access)
	ch := e.MustDo(200, "GET", "/conversations/"+cid+"/messages", nil, c.Access)["messages"].([]any)
	if len(ch) != 1 {
		t.Fatalf("new member sees %d messages, want only the 1 after joining", len(ch))
	}
	if c.readKey(t, e, cid, 1) != nil {
		t.Fatal("new member received the old key")
	}

	// delete: author may; others may not
	if code, _ := e.Do("DELETE", "/conversations/"+cid+"/messages/"+m0["id"].(string), nil, a.Access); code != 404 {
		// m0 is epoch 1 and A is a member: A is not the author and has no Moderate right
		if code != 403 {
			t.Fatalf("non-author deleted: %d", code)
		}
	}
	e.MustDo(204, "DELETE", "/conversations/"+cid+"/messages/"+m0["id"].(string), nil, owner.Access)
}

func owner2signing(t *testing.T, e *testkit.Env, viewer *member, person, deviceID string) []byte {
	t.Helper()
	for _, d := range e.MustDo(200, "GET", "/e2ee/people/"+person+"/devices", nil, viewer.Access)["devices"].([]any) {
		dm := d.(map[string]any)
		if dm["id"] == deviceID {
			b, _ := base64.StdEncoding.DecodeString(dm["signing_key"].(string))
			return b
		}
	}
	t.Fatal("sender device not in directory")
	return nil
}

func TestDirectMessagesAreOncePerPairAndRealtime(t *testing.T) {
	e := testkit.New(t)
	owner, slug := e.NewOrg("Olivia Owner")
	org := orgID(t, e, slug)
	o := &member{User: owner}
	o.person = e.MustDo(200, "GET", "/me", nil, owner.Access)["person"].(string)
	a := newMember(t, e, owner, slug, "Amy")
	o.register(t, e)
	a.register(t, e)
	key, _ := e2ee.NewConversationKey()
	body := map[string]any{"kind": "dm", "members": []string{a.person}, "keys": wrapFor(t, e, o, key, 1, "", o.person, a.person)}
	first := e.MustDo(201, "POST", "/conversations", body, owner.Access)
	again := e.MustDo(200, "POST", "/conversations", body, owner.Access)
	if first["id"] != again["id"] {
		t.Fatal("a second direct message was created for the same pair")
	}
	cid := first["id"].(string)

	// Amy listens in real time; the event carries ciphertext, never text
	tk := e.MustDo(200, "POST", "/realtime/ticket", nil, a.Access)["ticket"].(string)
	ws, _, err := websocket.Dial(context.Background(), "ws"+strings.TrimPrefix(e.Srv.URL, "http")+"/v1/realtime?ticket="+tk, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ws.CloseNow()
	_ = ws.Write(context.Background(), websocket.MessageText, []byte(`{"op":"sub","conv":"`+cid+`"}`))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for {
		_, b, err := ws.Read(ctx)
		if err != nil {
			t.Fatal("no subscription confirmation:", err)
		}
		if strings.Contains(string(b), `"subscribed"`) {
			break
		}
	}
	o.send(t, e, org, cid, 1, key, "hello Amy, this is private", 201)
	for {
		_, b, err := ws.Read(ctx)
		if err != nil {
			t.Fatal("message event never arrived:", err)
		}
		var ev map[string]any
		_ = json.Unmarshal(b, &ev)
		if ev["type"] == "message.new" {
			if strings.Contains(string(b), "hello Amy") {
				t.Fatal("event leaked the text")
			}
			break
		}
	}
	// nobody outside can start a conversation that includes someone not in the org
	if code, _ := e.Do("POST", "/conversations", map[string]any{"kind": "dm", "members": []string{uuid.NewString()}, "keys": []any{}}, owner.Access); code != 400 {
		t.Fatalf("dm with a stranger: %d", code)
	}
}

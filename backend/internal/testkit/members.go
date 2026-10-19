package testkit

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"regexp"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/e2ee"
	"hub/internal/modules/keys"
)

var tokenRe = regexp.MustCompile(`/join\?t=(\S+)`)

// Member is a person in an organization with a real end-to-end encryption device, acting like the app would.
type Member struct {
	*User
	Person string
	Acct   string
	Dev    *e2ee.Device
	DevID  string
}

func b64(b []byte) string { return base64.StdEncoding.EncodeToString(b) }

// Owner wraps the user NewOrg returns.
func (e *Env) Owner(u *User) *Member {
	m := &Member{User: u}
	me := e.MustDo(200, "GET", "/me", nil, u.Access)
	m.Person, m.Acct = me["person"].(string), me["account"].(string)
	return m
}

// Join onboards a person as the owner would and has them accept the emailed invitation.
func (e *Env) Join(owner *User, slug, first, role string) *Member {
	e.T.Helper()
	email := Uniq()
	body := map[string]any{"first_name": first, "last_name": "T", "email": email, "phone": "+254700000001", "date_of_birth": "1990-04-05", "gender": "Male"}
	if role != "" {
		body["role"] = role
	}
	out := e.MustDo(201, "POST", "/people", map[string]any{"people": []any{body}}, owner.Access)
	pid := out["results"].([]any)[0].(map[string]any)["id"].(string)
	tok := tokenRe.FindStringSubmatch(e.Dev.To(email)[0].Body)[1]
	e.Signup(first, email)
	e.MustDo(200, "POST", "/invites/accept", map[string]any{"token": tok}, e.Login(email, "").Access)
	m := e.Owner(e.Login(email, slug))
	if m.Person != pid {
		e.T.Fatalf("joined as the wrong person")
	}
	return m
}

// Register creates a device and publishes its public keys.
func (m *Member) Register(e *Env) {
	e.T.Helper()
	d, err := e2ee.NewDevice()
	if err != nil {
		e.T.Fatal(err)
	}
	m.Dev = d
	acct, _ := uuid.Parse(m.Acct)
	pub := d.Public()
	bind := ed25519.Sign(d.Sign, keys.BindingMessage(pub.Identity, acct))
	m.DevID = e.MustDo(201, "POST", "/e2ee/devices", map[string]any{"label": "test", "identity_key": b64(pub.Identity), "signing_key": b64(pub.Signing), "binding_sig": b64(bind)}, m.Access)["id"].(string)
}

// WrapFor wraps a conversation key for every active device of the given people.
func (m *Member) WrapFor(e *Env, key []byte, epoch int, people ...string) []map[string]any {
	e.T.Helper()
	var out []map[string]any
	for _, pid := range people {
		for _, d := range e.MustDo(200, "GET", "/e2ee/people/"+pid+"/devices", nil, m.Access)["devices"].([]any) {
			dm := d.(map[string]any)
			ik, _ := base64.StdEncoding.DecodeString(dm["identity_key"].(string))
			w, err := e2ee.Wrap(key, ik, []byte("ctx"))
			if err != nil {
				e.T.Fatal(err)
			}
			out = append(out, map[string]any{"epoch": epoch, "device": dm["id"], "wrapped_key": b64(w)})
		}
	}
	return out
}

// Seal encrypts text as this member for a conversation.
func (m *Member) Seal(org, conv, id string, epoch int, key []byte, text string) string {
	env, err := m.Dev.Seal(key, []byte(text), e2ee.Context{Org: org, Conversation: conv, Sender: m.Acct, MessageID: id, Epoch: epoch})
	if err != nil {
		panic(err)
	}
	return b64(env)
}

func (m *Member) Send(e *Env, org, conv string, epoch int, key []byte, text string, want int) map[string]any {
	e.T.Helper()
	id := uuid.Must(uuid.NewV7()).String()
	return e.MustDo(want, "POST", "/conversations/"+conv+"/messages", map[string]any{"id": id, "epoch": epoch, "device": m.DevID, "envelope": m.Seal(org, conv, id, epoch, key, text)}, m.Access)
}

func (m *Member) ReadKey(e *Env, conv string, epoch int) []byte {
	e.T.Helper()
	for _, k := range e.MustDo(200, "GET", "/conversations/"+conv+"/keys", nil, m.Access)["keys"].([]any) {
		km := k.(map[string]any)
		if int(km["epoch"].(float64)) == epoch && km["device"] == m.DevID {
			w, _ := base64.StdEncoding.DecodeString(km["wrapped_key"].(string))
			key, err := m.Dev.Unwrap(w, []byte("ctx"))
			if err != nil {
				e.T.Fatal(err)
			}
			return key
		}
	}
	return nil
}

func (e *Env) OrgID(slug string) string {
	var id string
	_ = e.A.DB.Tx(context.Background(), func(tx pgx.Tx) error {
		return tx.QueryRow(context.Background(), `select id::text from hub.orgs where slug=$1`, slug).Scan(&id)
	})
	return id
}

// MakeOverdue pretends a ticket's TAT ran out an hour ago. It writes as a superuser to simulate the passing of time,
// which is the only way to do that without waiting; the integrity scanner would flag it in a real database.
func (e *Env) MakeOverdue(conv string) {
	e.T.Helper()
	root, err := pgx.Connect(context.Background(), "postgres://hubadmin@/hub?host=/tmp&port=5433")
	if err != nil {
		e.T.Skip("superuser not reachable: ", err)
	}
	defer root.Close(context.Background())
	_, _ = root.Exec(context.Background(), "set session_replication_role = replica")
	if _, err := root.Exec(context.Background(), `update hub.tickets set due_at = now() - interval '1 hour', created_at = now() - interval '5 hours' where conversation_id=$1`, conv); err != nil {
		e.T.Fatal(err)
	}
}

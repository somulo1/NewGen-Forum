// Package seeder loads the demo organization: the frontend's mock data, played back through the real services so every
// rule, approval, encryption step and audit entry is the one live code uses. Encrypted content is made by "demo clients":
// each demo person has a device whose keys are derived from a seed, so the seeded chats, posts and files are real
// ciphertext that those devices can open. That makes the demo readable by anyone who knows the seed. It is for development
// and testing only, and the seeder refuses to run when HUB_ENV=prod.
package seeder

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/e2ee"
	"hub/internal/mail"
	"hub/internal/modules/approvals"
	"hub/internal/modules/auth"
	"hub/internal/modules/calls"
	"hub/internal/modules/chat"
	"hub/internal/modules/comms"
	"hub/internal/modules/files"
	"hub/internal/modules/keys"
	"hub/internal/modules/org"
	"hub/internal/modules/people"
	"hub/internal/modules/projects"
	"hub/internal/modules/roles"
	"hub/internal/modules/security"
	"hub/internal/modules/settings"
	"hub/internal/modules/tickets"
	"hub/internal/modules/tones"
	"hub/internal/tenant"
	"hub/seed"
)

const (
	DemoPassword = "Hub!Demo2026"
	OwnerEmail   = "samwel.omulo@company.com"
	OwnerName    = "Samwel Okoth Omulo"
	OrgSlug      = "demo"
)

// Actor is one demo person with a working device.
type Actor struct {
	Name, First, Last, Email, Title, Color, Initials, Availability string
	Role                                                           string
	Account, Person                                                uuid.UUID
	Dev                                                            *e2ee.Device
	DevID                                                          uuid.UUID
	P                                                              *tenant.Principal
	deptID                                                         uuid.UUID
}

type Seeder struct {
	A    *app.App
	Dev  *mail.Dev
	Auth *auth.Service
	Ap   *approvals.Engine
	Org  *org.Service
	Ppl  *people.Service
	Rol  *roles.Service
	Keys *keys.Module
	Chat *chat.Service
	Tick *tickets.Service
	Proj *projects.Service
	Com  *comms.Service
	File *files.Service
	Sec  *security.Module
	Set  *settings.Module
	Call *calls.Module
	Tone *tones.Module

	Secret  string
	OrgID   uuid.UUID
	Owner   *Actor
	ByEmail map[string]*Actor
	ByName  map[string]*Actor // first actor with that name (the mock data has people who share a name)
	All     []*Actor
	Units   map[string]uuid.UUID // "kind:name" -> id
	UnitIDs map[string]uuid.UUID // mock id (s1, b2, u3, d4) -> id
	ConvKey map[uuid.UUID][]byte
	ConvEp  map[uuid.UUID]int
	Now     time.Time
	Log     *slog.Logger
	forums  []forumRef
	folder  map[string]uuid.UUID
	wall    uuid.UUID

	data struct {
		state  map[string]json.RawMessage
		dir    []dirPerson
		org    orgData
		logs   []baseLog
		forums []forumData
	}
}

type dirPerson struct {
	ID                                                                   int
	Name, Email, Initials, Dept, Title, Loc, Sector, Unit, Color, Status string
}

func New(a *app.App, dev *mail.Dev, secret string) *Seeder {
	au := auth.New(a)
	ap := approvals.New(a)
	s := &Seeder{A: a, Dev: dev, Auth: au, Ap: ap, Secret: secret, ByEmail: map[string]*Actor{}, ByName: map[string]*Actor{}, Units: map[string]uuid.UUID{}, UnitIDs: map[string]uuid.UUID{},
		ConvKey: map[uuid.UUID][]byte{}, ConvEp: map[uuid.UUID]int{}, Now: time.Now().UTC(), Log: slog.Default()}
	s.Org = org.New(a, au)
	s.Ppl = people.New(a, au, ap)
	s.Rol = roles.New(a, ap)
	s.Keys = keys.New(a)
	s.Chat = chat.New(a)
	s.Tick = tickets.New(a)
	s.Proj = projects.New(a, ap)
	s.Com = comms.New(a, ap)
	s.File = files.New(a, ap)
	s.Sec = security.New(a)
	s.Set = settings.New(a)
	s.Call = calls.New(a)
	s.Tone = tones.New(a)
	return s
}

func (s *Seeder) load() error {
	read := func(name string, dst any) error {
		b, err := seed.FS.ReadFile("data/" + name)
		if err != nil {
			return err
		}
		return json.Unmarshal(b, dst)
	}
	if err := read("state.json", &s.data.state); err != nil {
		return err
	}
	for name, dst := range map[string]any{"dir.json": &s.data.dir, "org.json": &s.data.org, "baseLogs.json": &s.data.logs, "forums.json": &s.data.forums} {
		if err := read(name, dst); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}
	return nil
}

// st decodes one top-level key of the exported mock state.
func (s *Seeder) st(key string, dst any) {
	raw, ok := s.data.state[key]
	if !ok {
		panic("seed: state has no " + key)
	}
	if err := json.Unmarshal(raw, dst); err != nil {
		panic(fmt.Sprintf("seed: %s: %v", key, err))
	}
}

func (s *Seeder) ago(d time.Duration) time.Time { return s.Now.Add(-d).Truncate(time.Microsecond) }

func (s *Seeder) deviceFor(email string) (*e2ee.Device, error) {
	m := hmac.New(sha256.New, []byte(s.Secret))
	m.Write([]byte("hub-demo-device|" + email))
	return e2ee.DeviceFromSeed(m.Sum(nil))
}

// DeviceSeed shows how a demo person's device keys are derived, so a development client can sign in as them.
func DeviceSeed(secret, email string) string {
	m := hmac.New(sha256.New, []byte(secret))
	m.Write([]byte("hub-demo-device|" + email))
	return base64.StdEncoding.EncodeToString(m.Sum(nil))
}

func (s *Seeder) must(err error, what string) {
	if err != nil {
		panic(fmt.Sprintf("seed: %s: %v", what, err))
	}
}

// Run loads everything. It is safe to run once on an empty database.
func (s *Seeder) Run(ctx context.Context) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	if s.A.Cfg.Env == "prod" {
		return fmt.Errorf("the demo seeder never runs in production")
	}
	if err := s.load(); err != nil {
		return err
	}
	steps := []struct {
		name string
		fn   func(context.Context)
	}{
		{"tones", s.seedTones},
		{"organization and people", s.seedPeople},
		{"settings and rules", s.seedSettings},
		{"direct messages", s.seedDMs},
		{"forums and polls", s.seedForums},
		{"community wall and moments", s.seedWall},
		{"notes", s.seedNotes},
		{"files", s.seedFiles},
		{"tickets", s.seedTickets},
		{"projects", s.seedProjects},
		{"calendar", s.seedCalendar},
		{"communications, awards and approvals", s.seedComms},
		{"calls", s.seedCalls},
		{"security log and alerts", s.seedSecurity},
		{"notifications and preferences", s.seedNotifications},
	}
	for _, st := range steps {
		t0 := time.Now()
		st.fn(ctx)
		s.Log.Info("seeded", "step", st.name, "took", time.Since(t0).Round(time.Millisecond))
	}
	return nil
}

// ---- shared helpers ----

func (s *Seeder) principal(a *Actor) *tenant.Principal {
	return &tenant.Principal{Account: a.Account, Org: s.OrgID, Person: a.Person, Name: a.Name, IP: "10.0.4.21"}
}

func (s *Seeder) actor(name string) *Actor {
	n := strings.TrimSpace(name)
	switch n {
	case "Sam", "me", "You", "Samwel Okoth Omulo":
		return s.Owner
	}
	if a, ok := s.ByName[n]; ok {
		return a
	}
	// "Brian K." style: first name + last initial
	if i := strings.LastIndex(n, " "); i > 0 && strings.HasSuffix(n, ".") {
		first, li := n[:i], strings.TrimSuffix(n[i+1:], ".")
		for _, a := range s.All {
			if a.First == first && strings.HasPrefix(a.Last, li) {
				return a
			}
		}
	}
	panic("seed: no demo person called " + name)
}

func (s *Seeder) keyWraps(conv uuid.UUID, key []byte, epoch int, members ...*Actor) []chat.KeyWrap {
	var out []chat.KeyWrap
	for _, m := range members {
		w, err := e2ee.Wrap(key, m.Dev.Public().Identity, []byte("ctx"))
		s.must(err, "wrap key")
		out = append(out, chat.KeyWrap{Epoch: epoch, Device: m.DevID, Wrapped: base64.StdEncoding.EncodeToString(w)})
	}
	return out
}

// newConversation creates a container through the real service, with a fresh key wrapped for every member's device.
func (s *Seeder) newConversation(ctx context.Context, creator *Actor, in chat.CreateIn, members []*Actor) (*chat.Conversation, []byte) {
	key, err := e2ee.NewConversationKey()
	s.must(err, "conversation key")
	all := append([]*Actor{creator}, members...)
	seen := map[uuid.UUID]bool{}
	uniq := all[:0:0]
	for _, m := range all {
		if !seen[m.Person] {
			seen[m.Person] = true
			uniq = append(uniq, m)
		}
	}
	in.Keys = s.keyWraps(uuid.Nil, key, 1, uniq...)
	if in.Audience == "" {
		in.Members = nil
		for _, m := range uniq[1:] {
			in.Members = append(in.Members, m.Person)
		}
	}
	c, _, err := s.Chat.Create(ctx, s.principal(creator), in)
	s.must(err, "create "+in.Kind+" "+in.Name)
	s.ConvKey[c.ID], s.ConvEp[c.ID] = key, 1
	return c, key
}

// say puts one encrypted message from a demo person into a conversation at a given time, the way their device would have.
func (s *Seeder) say(ctx context.Context, conv uuid.UUID, from *Actor, typ string, payload any, at time.Time, replyTo *uuid.UUID) uuid.UUID {
	id := idAt(at)
	b, err := json.Marshal(payload)
	s.must(err, "payload")
	env, err := from.Dev.Seal(s.ConvKey[conv], b, e2ee.Context{Org: s.OrgID.String(), Conversation: conv.String(), Sender: from.Account.String(), MessageID: id.String(), Epoch: s.ConvEp[conv]})
	s.must(err, "seal")
	at = at.Truncate(time.Microsecond)
	err = s.A.DB.WithOrg(ctx, s.OrgID, func(tx pgx.Tx) error {
		return chat.PostMessage(ctx, tx, s.A, chat.Posted{Org: s.OrgID, Conversation: conv, ID: id, Account: from.Account, Person: from.Person, Device: from.DevID,
			Epoch: s.ConvEp[conv], Envelope: env, Type: typ, ReplyTo: replyTo, CreatedAt: &at})
	})
	s.must(err, "post message")
	return id
}

// sealRecord returns a base64 envelope for a record or file description made by a demo person.
func (s *Seeder) sealFor(conv uuid.UUID, from *Actor, id uuid.UUID, payload any) string {
	b, err := json.Marshal(payload)
	s.must(err, "payload")
	env, err := from.Dev.Seal(s.ConvKey[conv], b, e2ee.Context{Org: s.OrgID.String(), Conversation: conv.String(), Sender: from.Account.String(), MessageID: id.String(), Epoch: s.ConvEp[conv]})
	s.must(err, "seal")
	return base64.StdEncoding.EncodeToString(env)
}

// idAt makes a time-ordered UUID (v7) carrying the given moment, so message order matches the story being told.
func idAt(t time.Time) uuid.UUID {
	id := uuid.Must(uuid.NewV7())
	ms := uint64(t.UnixMilli())
	id[0], id[1], id[2], id[3], id[4], id[5] = byte(ms>>40), byte(ms>>32), byte(ms>>24), byte(ms>>16), byte(ms>>8), byte(ms)
	return id
}

// markUnread leaves the owner with exactly n unread messages from others in a conversation.
func (s *Seeder) markUnread(ctx context.Context, conv uuid.UUID, ids []uuid.UUID, fromOthers []bool, unread int) {
	// walk back from the newest message, counting messages from others; the cursor sits just before the n-th of them
	seen, cut := 0, -1
	for i := len(ids) - 1; i >= 0; i-- {
		if fromOthers[i] {
			seen++
			if seen == unread {
				cut = i - 1
				break
			}
		}
	}
	var upTo uuid.UUID
	if unread == 0 {
		upTo = ids[len(ids)-1]
	} else if cut >= 0 {
		upTo = ids[cut]
	} else {
		return
	}
	s.must(s.Chat.MarkRead(ctx, s.principal(s.Owner), conv, upTo), "mark read")
}

type pgxTx = pgx.Tx

func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

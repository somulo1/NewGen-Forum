package calls

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"

	"hub/internal/authz"
	"hub/internal/callsvc"
	"hub/internal/httpx"
	"hub/internal/sfu"
	"hub/internal/tenant"
)

// Big rooms run on media servers (SFU). A room is a "stage" of people who speak and a crowd who listen: everyone's audio and
// video goes to the media server once, and the server sends each listener only the few loudest speakers. Private calls (two
// people) and installs without a media server keep using the small peer-to-peer rooms on the calls service.
//
// Admission is counted in Redis so every API node sees the same numbers, and the media server enforces the same cap again
// (a hard limit on people per room) in case a count is ever wrong. Counts are sorted sets scored by when a reservation
// lapses: a pass that is never used frees its place after two minutes; a person who joined holds it until they leave.

const forever = 4e12 // a score far in the future (seconds)

type SFUInfo struct {
	URL      string `json:"url"`
	Token    string `json:"token"`
	Room     string `json:"room"`
	Identity string `json:"identity"`
	Speak    bool   `json:"can_speak"` // may publish audio/video right now
	Limit    int    `json:"limit"`     // most people in this room
	Stage    int    `json:"stage"`     // most people who may speak at once in a big room
	OpenMic  int    `json:"open_mic"`  // rooms up to this size: everybody speaks
	Role     string `json:"role"`
}

func (m *Module) sfu() *sfu.Client {
	if m.sfuc == nil {
		m.sfuc = sfu.New(m.A.Cfg.SFUNodes)
	}
	return m.sfuc
}

func (m *Module) sfuFor(context_ string) bool { return m.sfu().On() && context_ != "dm" }

func roomName(org, call uuid.UUID) string { return org.String() + "." + call.String() }

func parseRoom(n string) (org, call uuid.UUID, ok bool) {
	o, c, found := strings.Cut(n, ".")
	var e1, e2 error
	org, e1 = uuid.Parse(o)
	call, e2 = uuid.Parse(c)
	return org, call, found && e1 == nil && e2 == nil
}

func (m *Module) limitFor(kind string) int {
	if kind == "video" {
		return m.A.Cfg.CallLimits.Video
	}
	return m.A.Cfg.CallLimits.Voice
}

func sources(kind string) []string {
	if kind == "video" {
		return []string{"microphone", "camera", "screen_share", "screen_share_audio"}
	}
	return []string{"microphone"}
}

func now() float64 { return float64(time.Now().Unix()) }

// count drops lapsed reservations and returns how many places are held.
func (m *Module) held(ctx context.Context, key string) int64 {
	m.A.Rdb.ZRemRangeByScore(ctx, key, "-inf", fmt.Sprint(now()))
	n, _ := m.A.Rdb.ZCard(ctx, key).Result()
	return n
}

// reserveRoom claims a place among the rooms running now, for the organization and for the whole deployment.
func (m *Module) reserveRoom(ctx context.Context, org, call uuid.UUID) error {
	lim := m.A.Cfg.CallLimits
	ok, ko := "sfu:r:"+org.String(), "sfu:rooms"
	m.A.Rdb.ZAdd(ctx, ok, redis.Z{Score: now() + 300, Member: call.String()})
	m.A.Rdb.ZAdd(ctx, ko, redis.Z{Score: now() + 300, Member: call.String()})
	if n := m.held(ctx, ok); int(n) > lim.RoomsOrg {
		m.releaseRoom(ctx, org, call)
		return httpx.Err(429, "too_many_calls", fmt.Sprintf("Your organization is already running %d calls, the most it can run at once. End one and try again.", lim.RoomsOrg))
	}
	if n := m.held(ctx, ko); int(n) > lim.RoomsTotal {
		m.releaseRoom(ctx, org, call)
		return httpx.Err(503, "calls_busy", "Hub is carrying as many calls as it can right now. Try again in a moment.")
	}
	return nil
}

func (m *Module) releaseRoom(ctx context.Context, org, call uuid.UUID) {
	m.A.Rdb.ZRem(ctx, "sfu:r:"+org.String(), call.String())
	m.A.Rdb.ZRem(ctx, "sfu:rooms", call.String())
}

// sfuEntry opens the room on its media server (when new) and returns the pass for one person.
func (m *Module) sfuEntry(ctx context.Context, p *tenant.Principal, c *Call, role string, create bool) (*SFUInfo, error) {
	org := p.Org
	room := roomName(org, c.ID)
	lim := m.A.Cfg.CallLimits
	max := m.limitFor(c.Kind)
	rc := m.A.Rdb
	if role == "participant" {
		if ok, _ := rc.Exists(ctx, "sfu:lock:"+room).Result(); ok > 0 {
			return nil, httpx.Err(423, "room_locked", "The host locked this call.")
		}
	}
	// a place in the room, counted before the pass is handed out
	identity := p.Person.String() + "." + strings.ReplaceAll(uuid.NewString(), "-", "")[:6]
	key := "sfu:p:" + room
	held := m.held(ctx, key)
	rc.ZAdd(ctx, key, redis.Z{Score: now() + 120, Member: identity})
	rc.Expire(ctx, key, 24*time.Hour)
	if int(held)+1 > max {
		rc.ZRem(ctx, key, identity)
		return nil, httpx.Err(409, "room_full", fmt.Sprintf("This call is full. %d people is the most a %s call can hold.", max, c.Kind))
	}
	fail := func(err error) (*SFUInfo, error) { rc.ZRem(ctx, key, identity); return nil, err }
	if create {
		if err := m.sfu().CreateRoom(ctx, room, max, 5*time.Minute); err != nil {
			return fail(httpx.Err(503, "calls_unavailable", "Calls are busy right now. Try again in a moment."))
		}
	}
	// everyone speaks in a small room; in a big one only hosts, moderators and people brought on stage do
	speak := role != "participant" || int(held)+1 <= lim.OpenMic
	if !speak {
		if ok, _ := rc.SIsMember(ctx, "sfu:stage:"+room, p.Person.String()).Result(); ok {
			speak = true
		}
	}
	meta, _ := json.Marshal(map[string]string{"person": p.Person.String(), "role": role})
	node := m.sfu().Pick(room)
	tok, err := sfu.Token(node, sfu.Grant{Room: room, Identity: identity, Name: p.Name, Metadata: string(meta), Publish: speak, Sources: sources(c.Kind)})
	if err != nil {
		return fail(err)
	}
	return &SFUInfo{URL: sfu.WS(node), Token: tok, Room: room, Identity: identity, Speak: speak, Limit: max, Stage: lim.Stage, OpenMic: lim.OpenMic, Role: role}, nil
}

// ---- room controls ----

type ModIn struct {
	Action   string `json:"action"`   // mute | cam_off | remove | mute_all | lock | unlock | end | stage_on | stage_off
	Identity string `json:"identity"` // the target's id in the room
}

// Moderate lets the host or a moderator control a big room. The media server carries it out.
func (m *Module) Moderate(ctx context.Context, p *tenant.Principal, id uuid.UUID, in ModIn) error {
	var c *Call
	var role string
	err := m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		var err error
		if c, err = m.load(ctx, tx, id, p.Person); err != nil {
			return err
		}
		role = m.roleOf(ctx, tx, c, p)
		return nil
	})
	if err != nil {
		return err
	}
	if !m.sfu().On() || c.Context == "dm" {
		return httpx.Bad("This call has no room controls.")
	}
	if role == "participant" {
		return httpx.Err(403, "forbidden", "Only the host or a moderator can do that.")
	}
	room := roomName(p.Org, c.ID)
	hostID := c.Host.String()
	target := strings.SplitN(in.Identity, ".", 2)[0]
	if (in.Action == "remove" || in.Action == "mute" || in.Action == "cam_off" || in.Action == "stage_off") && target == hostID && p.Person != c.Host {
		return httpx.Err(403, "forbidden", "Only the host can do that to the host.")
	}
	cl := m.sfu()
	rc := m.A.Rdb
	switch in.Action {
	case "mute":
		return cl.Mute(ctx, room, in.Identity, true)
	case "cam_off":
		return cl.Mute(ctx, room, in.Identity, false)
	case "remove":
		return cl.RemoveParticipant(ctx, room, in.Identity)
	case "mute_all":
		ps, err := cl.List(ctx, room)
		if err != nil {
			return err
		}
		keep := map[string]bool{}
		for _, x := range ps {
			if strings.HasPrefix(x.Identity, hostID+".") || strings.HasPrefix(x.Identity, p.Person.String()+".") {
				keep[x.Identity] = true
			}
		}
		return cl.MuteAll(ctx, room, keep)
	case "lock":
		return rc.Set(ctx, "sfu:lock:"+room, 1, 24*time.Hour).Err()
	case "unlock":
		return rc.Del(ctx, "sfu:lock:"+room).Err()
	case "end":
		if p.Person != c.Host {
			return httpx.Err(403, "forbidden", "Only the host can end the call for everyone.")
		}
		return cl.DeleteRoom(ctx, room)
	case "stage_on":
		ps, err := cl.List(ctx, room)
		if err != nil {
			return err
		}
		speaking := 0
		for _, x := range ps {
			if x.Permission.CanPublish {
				speaking++
			}
		}
		if speaking >= m.A.Cfg.CallLimits.Stage {
			return httpx.Err(409, "stage_full", fmt.Sprintf("%d people are already on stage, the most a call can have. Ask someone to step down first.", m.A.Cfg.CallLimits.Stage))
		}
		rc.SAdd(ctx, "sfu:stage:"+room, target)
		rc.Expire(ctx, "sfu:stage:"+room, 24*time.Hour)
		return cl.SetStage(ctx, room, in.Identity, true, sources(c.Kind))
	case "stage_off":
		rc.SRem(ctx, "sfu:stage:"+room, target)
		return cl.SetStage(ctx, room, in.Identity, false, nil)
	}
	return httpx.Bad("Unknown action.")
}

func (m *Module) roleOf(ctx context.Context, tx pgx.Tx, c *Call, p *tenant.Principal) string {
	if c.Host == p.Person {
		return "host"
	}
	if mod, _ := authz.Can(ctx, tx, p.Person, "forums", "Moderate"); mod {
		return "moderator"
	}
	return "participant"
}

// ---- notices from the media servers ----

// Hook receives the media server's signed notices (someone joined or left, the room closed). They keep the counts honest and
// feed the same event list as the peer-to-peer service, so history and ringing work the same either way.
func (m *Module) Hook(ctx context.Context, auth string, body []byte) error {
	e, err := m.sfu().Verify(auth, body)
	if err != nil {
		return httpx.Err(401, "bad_signature", "Not from a media server.")
	}
	org, call, ok := parseRoom(e.Room.Name)
	if !ok {
		return nil
	}
	rc := m.A.Rdb
	person := strings.SplitN(e.Participant.Identity, ".", 2)[0]
	push := func(typ string) {
		b, _ := json.Marshal(map[string]any{"type": typ, "call": call, "org": org, "person": person, "at": time.Now().UTC()})
		rc.RPush(ctx, callsvc.EventsKey, b)
	}
	switch e.Event {
	case "room_started":
		rc.ZAdd(ctx, "sfu:r:"+org.String(), redis.Z{Score: forever, Member: call.String()})
		rc.ZAdd(ctx, "sfu:rooms", redis.Z{Score: forever, Member: call.String()})
	case "participant_joined":
		rc.ZAdd(ctx, "sfu:p:"+e.Room.Name, redis.Z{Score: forever, Member: e.Participant.Identity})
		push("joined")
	case "participant_left":
		rc.ZRem(ctx, "sfu:p:"+e.Room.Name, e.Participant.Identity)
		push("left")
	case "room_finished":
		m.releaseRoom(ctx, org, call)
		rc.Del(ctx, "sfu:p:"+e.Room.Name, "sfu:stage:"+e.Room.Name, "sfu:lock:"+e.Room.Name)
		push("ended")
	}
	return nil
}

func (m *Module) sfuPublic(public chi.Router) {
	public.Post("/calls/sfu-hook", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		if !m.sfu().On() {
			return httpx.ErrNotFound
		}
		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			return httpx.Bad("Bad notice.")
		}
		if err := m.Hook(r.Context(), r.Header.Get("Authorization"), body); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
}

func (m *Module) sfuPrivate(private chi.Router, org func(func(http.ResponseWriter, *http.Request, *tenant.Principal) error) http.HandlerFunc) {
	private.Post("/calls/{id}/control", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			return httpx.ErrNotFound
		}
		var in ModIn
		if err := httpx.Decode(r, &in, 2<<10); err != nil {
			return err
		}
		if err := m.Moderate(r.Context(), p, id, in); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
}

package calls_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"hub/internal/config"
	"hub/internal/e2ee"
	"hub/internal/testkit"
)

// fakeMedia stands in for a LiveKit server: it records the room controls it is asked for.
type fakeMedia struct {
	mu       sync.Mutex
	calls    []string
	speaking int
	max      int
}

func (f *fakeMedia) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b, _ := io.ReadAll(r.Body)
	method := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
	f.mu.Lock()
	f.calls = append(f.calls, method)
	var in map[string]any
	_ = json.Unmarshal(b, &in)
	if method == "CreateRoom" {
		f.max = int(in["maxParticipants"].(float64))
	}
	n := f.speaking
	f.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	if method == "ListParticipants" {
		ps := []map[string]any{}
		for i := 0; i < n; i++ {
			ps = append(ps, map[string]any{"identity": "x.1", "permission": map[string]any{"canPublish": true}})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"participants": ps})
		return
	}
	_, _ = w.Write([]byte("{}"))
}

func (f *fakeMedia) saw(m string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, c := range f.calls {
		if c == m {
			return true
		}
	}
	return false
}

func TestBigRoomsOnAMediaServer(t *testing.T) {
	e := testkit.New(t)
	media := &fakeMedia{}
	ms := httptest.NewServer(media)
	defer ms.Close()
	node := config.SFUNode{URL: ms.URL, Key: "hubkey", Secret: "hubsecret-hubsecret-hubsecret-hubsecret"}
	e.A.Cfg.SFUNodes = []config.SFUNode{node}
	e.A.Cfg.CallLimits = config.CallLimits{Voice: 3, Video: 3, OpenMic: 2, Stage: 2, RoomsOrg: 1, RoomsTotal: 10}

	ownerU, slug := e.NewOrg("Olivia Owner")
	o := e.Owner(ownerU)
	a := e.Join(ownerU, slug, "Amy", "staff")
	b := e.Join(ownerU, slug, "Ben", "staff")
	c := e.Join(ownerU, slug, "Cy", "staff")
	for _, m := range []*testkit.Member{o, a, b, c} {
		m.Register(e)
	}
	key, _ := e2ee.NewConversationKey()
	conv := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "forum", "name": "Everyone", "members": []string{a.Person, b.Person, c.Person}, "keys": o.WrapFor(e, key, 1, o.Person, a.Person, b.Person, c.Person)}, ownerU.Access)["id"].(string)

	start := func() map[string]any {
		return e.MustDo(201, "POST", "/calls", map[string]any{"conversation": conv, "kind": "video", "context": "forum"}, ownerU.Access)
	}
	started := start()
	sf := started["sfu"].(map[string]any)
	if started["mode"] != "sfu" || sf["limit"].(float64) != 3 || !sf["can_speak"].(bool) {
		t.Fatalf("a forum call should run on the media server: %v", started)
	}
	if media.max != 3 || !media.saw("CreateRoom") {
		t.Fatalf("the room cap must be set on the media server too: %d", media.max)
	}
	// the pass is signed for that media server and names the room
	var cl struct {
		jwt.RegisteredClaims
		Video map[string]any `json:"video"`
	}
	if _, err := jwt.ParseWithClaims(sf["token"].(string), &cl, func(*jwt.Token) (any, error) { return []byte(node.Secret), nil }); err != nil || cl.Video["room"] != sf["room"] || cl.Issuer != "hubkey" {
		t.Fatalf("pass: %v %v", err, cl)
	}
	callID := started["call"].(map[string]any)["id"].(string)

	// only one room may run at a time in this organization
	if code, _ := e.Do("POST", "/calls", map[string]any{"conversation": conv, "kind": "voice", "context": "forum"}, ownerU.Access); code != 429 {
		t.Fatalf("a second room should be refused: %d", code)
	}

	// room for 3: the host, then two more join; one more is turned away
	join := func(m *testkit.Member) (int, map[string]any) { return e.Do("POST", "/calls/"+callID+"/join", nil, m.Access) }
	code, ja := join(a)
	if code != 200 || !ja["sfu"].(map[string]any)["can_speak"].(bool) {
		t.Fatalf("Amy: %d %v", code, ja)
	}
	code, jb := join(b)
	if code != 200 || jb["sfu"].(map[string]any)["can_speak"].(bool) {
		t.Fatalf("Ben joins as a listener once the open-mic size is passed: %d %v", code, jb)
	}
	if code, _ = join(c); code != 409 {
		t.Fatalf("a full room let someone in: %d", code)
	}

	// only the host or a moderator brings a listener on stage, and the stage has a cap
	ctl := func(m *testkit.Member, action, id string) int {
		code, _ := e.Do("POST", "/calls/"+callID+"/control", map[string]any{"action": action, "identity": id}, m.Access)
		return code
	}
	benID := jb["sfu"].(map[string]any)["identity"].(string)
	if code := ctl(a, "stage_on", benID); code != 403 {
		t.Fatalf("a participant ran a room control: %d", code)
	}
	media.mu.Lock()
	media.speaking = 2
	media.mu.Unlock()
	if code := ctl(o, "stage_on", benID); code != 409 {
		t.Fatalf("the stage is full: %d", code)
	}
	media.mu.Lock()
	media.speaking = 1
	media.mu.Unlock()
	if code := ctl(o, "stage_on", benID); code != 204 || !media.saw("UpdateParticipant") {
		t.Fatalf("stage_on: %d", code)
	}
	if code := ctl(o, "remove", benID); code != 204 || !media.saw("RemoveParticipant") {
		t.Fatalf("remove: %d", code)
	}
	// nobody can remove the host except the host
	hostID := sf["identity"].(string)
	if code := ctl(a, "remove", hostID); code != 403 {
		t.Fatalf("removed the host: %d", code)
	}

	// a lock keeps newcomers out
	if code := ctl(o, "lock", ""); code != 204 {
		t.Fatalf("lock: %d", code)
	}
	if code, _ := join(b); code != 423 {
		t.Fatalf("a locked room let someone in: %d", code)
	}

	// notices from the media server must be signed by it
	body := []byte(`{"event":"participant_joined","room":{"name":"` + sf["room"].(string) + `"},"participant":{"identity":"` + hostID + `"}}`)
	sum := sha256.Sum256(body)
	sign := func(secret string) string {
		tok, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"iss": "hubkey", "exp": time.Now().Add(time.Minute).Unix(), "sha256": base64.StdEncoding.EncodeToString(sum[:])}).SignedString([]byte(secret))
		return tok
	}
	post := func(auth string, b []byte) int {
		req, _ := http.NewRequest("POST", e.Srv.URL+"/v1/calls/sfu-hook", bytes.NewReader(b))
		req.Header.Set("Authorization", auth)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		return res.StatusCode
	}
	if code := post(sign("someone-elses-secret-someone-elses-key"), body); code != 401 {
		t.Fatalf("a forged notice was accepted: %d", code)
	}
	if code := post(sign(node.Secret), append([]byte(" "), body...)); code != 401 {
		t.Fatalf("an altered notice was accepted: %d", code)
	}
	if code := post(sign(node.Secret), body); code != 204 {
		t.Fatalf("a genuine notice was refused: %d", code)
	}
	if sc, err := e.A.Rdb.ZScore(t.Context(), "sfu:p:"+sf["room"].(string), hostID).Result(); err != nil || sc < 1e12 {
		t.Fatalf("a person in the room must hold their place until they leave: %v %v", sc, err)
	}

	// the room closing frees its place among the running rooms
	fin := []byte(`{"event":"room_finished","room":{"name":"` + sf["room"].(string) + `"}}`)
	sum = sha256.Sum256(fin)
	if code := post(sign(node.Secret), fin); code != 204 {
		t.Fatalf("room_finished: %d", code)
	}
	e.MustDo(201, "POST", "/calls", map[string]any{"conversation": conv, "kind": "voice", "context": "forum"}, ownerU.Access)
}

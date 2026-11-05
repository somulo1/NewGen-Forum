// Package callsvc is the calls microservice: it carries the signaling for voice and video rooms and the room controls. It
// never sees media (devices connect to each other, encrypted with DTLS-SRTP) and needs no database; it keeps rooms in
// memory and reports call events to Redis. Running it apart from the API means a big meeting cannot slow the rest of Hub.
package callsvc

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"hub/internal/calltoken"
)

const (
	EventsKey   = "calls:events"
	maxMessage  = 64 << 10
	sendBuffer  = 64
	msgsPerSec  = 60
	roomIdle    = 2 * time.Minute
	maxPeersCap = 8 // peer-to-peer mesh: every device sends to every other. An SFU lifts this (see docs).
)

type Server struct {
	Secret  []byte
	Rdb     *redis.Client
	Origins []string
	// Grace is how long a device whose connection dropped (a network change, a tunnel, a sleeping laptop) keeps its place in the room
	// while it reconnects. Nobody is told it left during that time, and the media between devices carries on.
	Grace time.Duration
	mu    sync.Mutex
	rooms map[string]*Room
	peers atomic.Int64
}

func New(secret []byte, rdb *redis.Client, origins []string) *Server {
	// the browser's Origin header is matched against host names, so "https://hub.example.com/" must become "hub.example.com"
	hosts := make([]string, 0, len(origins))
	for _, o := range origins {
		o = strings.TrimRight(strings.TrimSpace(o), "/")
		o = strings.TrimPrefix(strings.TrimPrefix(o, "https://"), "http://")
		if o != "" {
			hosts = append(hosts, o)
		}
	}
	return &Server{Secret: secret, Rdb: rdb, Origins: hosts, Grace: 20 * time.Second, rooms: map[string]*Room{}}
}

type Room struct {
	ID     string
	Org    string
	Kind   string
	mu     sync.Mutex
	peers  map[string]*Peer
	locked bool
	ended  bool
}

type Peer struct {
	ID        string `json:"id"`
	Person    string `json:"person"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	Mic       bool   `json:"mic"`
	Cam       bool   `json:"cam"`
	Share     bool   `json:"sharing"`
	Hand      bool   `json:"hand"`
	NetPaused bool   `json:"net_paused"`
	send      chan []byte
	cancel    context.CancelFunc
	// resuming a dropped connection: a secret only this device has (given over its own connection), which connection of the peer is
	// current, and whether it was removed on purpose (then it cannot come back)
	key      string
	gen      int
	detached bool
	final    bool
	timer    *time.Timer
	done     chan struct{} // closed when the current connection's writer has stopped
	cmu      sync.Mutex    // guards cancel, which is swapped when the device resumes
}

type event struct {
	Type   string    `json:"type"` // joined | left | ended
	Call   string    `json:"call"`
	Org    string    `json:"org"`
	Person string    `json:"person,omitempty"`
	At     time.Time `json:"at"`
}

func (s *Server) report(ctx context.Context, e event) {
	e.At = time.Now().UTC()
	b, _ := json.Marshal(e)
	if err := s.Rdb.RPush(ctx, EventsKey, b).Err(); err != nil {
		slog.Warn("could not report call event", "err", err)
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		s.mu.Lock()
		n := len(s.rooms)
		s.mu.Unlock()
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "rooms": n, "peers": s.peers.Load()})
	})
	mux.HandleFunc("GET /ws", s.ws)
	// metrics for the monitoring system: scrape from inside the network only
	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		n := len(s.rooms)
		s.mu.Unlock()
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		fmt.Fprintf(w, "# TYPE hub_calls_rooms gauge\nhub_calls_rooms %d\n# TYPE hub_calls_peers gauge\nhub_calls_peers %d\n", n, s.peers.Load())
	})
	return mux
}

type msg struct {
	Type      string          `json:"type"`
	To        string          `json:"to,omitempty"`
	From      string          `json:"from,omitempty"`
	Action    string          `json:"action,omitempty"`
	Target    string          `json:"target,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
	Mic       *bool           `json:"mic,omitempty"`
	Cam       *bool           `json:"cam,omitempty"`
	Share     *bool           `json:"sharing,omitempty"`
	Hand      *bool           `json:"hand,omitempty"`
	NetPaused *bool           `json:"net_paused,omitempty"`
}

func (s *Server) ws(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("resume") != "" {
		s.resume(w, r)
		return
	}
	claims, err := calltoken.Verify(s.Secret, r.URL.Query().Get("token"))
	if err != nil {
		slog.Warn("call room refused: invalid or expired pass", "err", err)
		http.Error(w, `{"error":"invalid or expired call pass"}`, http.StatusUnauthorized)
		return
	}
	// a pass works once
	if ok, err := s.Rdb.SetNX(r.Context(), "callpass:"+claims.ID, 1, 10*time.Minute).Result(); !ok && claims.ID != "" {
		slog.Warn("call room refused: pass already used or the pass store is unreachable", "err", err)
		http.Error(w, `{"error":"this call pass was already used"}`, http.StatusUnauthorized)
		return
	}
	room := s.room(claims)
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	kb := make([]byte, 16)
	_, _ = rand.Read(kb)
	peer := &Peer{ID: uuid.NewString(), Person: claims.Person, Name: claims.Name, Role: claims.Role, Mic: true, Cam: claims.Kind == "video", send: make(chan []byte, sendBuffer), cancel: cancel, key: hex.EncodeToString(kb)}
	if err := room.add(peer, claims.MaxPeer); err != nil {
		slog.Warn("call room refused", "err", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusConflict)
		return
	}
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: s.Origins})
	if err != nil {
		room.remove(peer)
		return
	}
	conn.SetReadLimit(maxMessage)
	s.peers.Add(1)
	s.report(ctx, event{Type: "joined", Call: room.ID, Org: room.Org, Person: peer.Person})

	room.mu.Lock()
	others := make([]*Peer, 0, len(room.peers))
	for _, p := range room.peers {
		if p.ID != peer.ID {
			others = append(others, p)
		}
	}
	gen := peer.gen
	room.mu.Unlock()
	welcome, _ := json.Marshal(map[string]any{"type": "welcome", "you": peer, "peers": others, "room": room.ID, "kind": room.Kind, "locked": room.locked, "resume": peer.key})
	peer.send <- welcome
	room.broadcast(peer.ID, map[string]any{"type": "peer-joined", "peer": peer})

	s.run(ctx, cancel, conn, room, peer, gen)
}

// resume reattaches a device whose connection dropped to the place it kept in the room. It proves it is that device with the secret it
// was given when it joined; no new pass is needed and nobody in the room hears that it ever left.
func (s *Server) resume(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	s.mu.Lock()
	room := s.rooms[q.Get("room")]
	s.mu.Unlock()
	var peer *Peer
	if room != nil {
		room.mu.Lock()
		if p := room.peers[q.Get("resume")]; p != nil && !p.final && !room.ended && subtle.ConstantTimeCompare([]byte(p.key), []byte(q.Get("key"))) == 1 {
			peer = p
		}
		room.mu.Unlock()
	}
	if peer == nil {
		http.Error(w, `{"error":"this place in the call is gone"}`, http.StatusGone)
		return
	}
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: s.Origins})
	if err != nil {
		return
	}
	conn.SetReadLimit(maxMessage)
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	// take over from the previous connection: stop it, and wait for its writer to let go of the outgoing queue
	room.mu.Lock()
	oldDone := peer.done
	peer.gen++
	gen := peer.gen
	peer.detached = false
	if peer.timer != nil {
		peer.timer.Stop()
		peer.timer = nil
	}
	room.mu.Unlock()
	peer.cmu.Lock()
	oldCancel := peer.cancel
	peer.cancel = cancel
	peer.cmu.Unlock()
	if oldCancel != nil {
		oldCancel()
	}
	if oldDone != nil {
		select {
		case <-oldDone:
		case <-time.After(3 * time.Second):
		}
	}
	s.peers.Add(1)
	room.mu.Lock()
	others := make([]*Peer, 0, len(room.peers))
	for _, p := range room.peers {
		if p.ID != peer.ID {
			others = append(others, p)
		}
	}
	room.mu.Unlock()
	resumed, _ := json.Marshal(map[string]any{"type": "resumed", "you": peer, "peers": others, "room": room.ID, "kind": room.Kind, "locked": room.locked})
	// what was queued while the device was away (offers, candidates) follows, after this
	pending := len(peer.send)
	queued := make([][]byte, 0, pending)
	for i := 0; i < pending; i++ {
		select {
		case b := <-peer.send:
			queued = append(queued, b)
		default:
		}
	}
	peer.send <- resumed
	for _, b := range queued {
		select {
		case peer.send <- b:
		default:
		}
	}
	s.run(ctx, cancel, conn, room, peer, gen)
}

// run carries one connection of a device: a writer, and a reader that handles what the device sends. When the connection ends, the
// device is removed at once if it left on purpose (it closed the connection properly), and otherwise keeps its place for the grace
// period so it can come back.
func (s *Server) run(ctx context.Context, cancel context.CancelFunc, conn *websocket.Conn, room *Room, peer *Peer, gen int) {
	done := make(chan struct{})
	room.mu.Lock()
	peer.done = done
	room.mu.Unlock()
	go func() { // writer
		defer close(done)
		t := time.NewTicker(10 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case b := <-peer.send:
				wc, c := context.WithTimeout(ctx, 8*time.Second)
				err := conn.Write(wc, websocket.MessageText, b)
				c()
				if err != nil {
					cancel()
					return
				}
			case <-t.C:
				pc, c := context.WithTimeout(ctx, 5*time.Second)
				err := conn.Ping(pc)
				c()
				if err != nil {
					cancel()
					return
				}
			}
		}
	}()

	var readErr error
	bucket, last := float64(msgsPerSec), time.Now()
	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			readErr = err
			break
		}
		// flood guard: a token bucket per connection
		now := time.Now()
		bucket = min(float64(msgsPerSec), bucket+now.Sub(last).Seconds()*msgsPerSec)
		last = now
		if bucket < 1 {
			continue
		}
		bucket--
		var m msg
		if json.Unmarshal(data, &m) != nil {
			continue
		}
		s.handle(room, peer, &m, data)
	}
	cancel()
	<-done
	s.peers.Add(-1)
	status := websocket.CloseStatus(readErr)
	leftOnPurpose := status == websocket.StatusNormalClosure || status == websocket.StatusGoingAway || status == websocket.StatusNoStatusRcvd
	s.detachOrLeave(room, peer, gen, leftOnPurpose)
	_ = conn.Close(websocket.StatusNormalClosure, "bye")
}

// detachOrLeave ends a connection's hold on the peer. A device that closed properly (hung up, closed the tab) is gone at once; one that
// dropped (the network went) keeps its place for Grace. A newer connection of the same device (gen) has already taken over: then
// this is only the old one finishing, and nothing changes.
func (s *Server) detachOrLeave(room *Room, peer *Peer, gen int, leftOnPurpose bool) {
	room.mu.Lock()
	if peer.gen != gen {
		room.mu.Unlock()
		return
	}
	if leftOnPurpose || peer.final || room.ended || s.Grace <= 0 {
		room.mu.Unlock()
		s.finish(room, peer)
		return
	}
	peer.detached = true
	peer.timer = time.AfterFunc(s.Grace, func() {
		room.mu.Lock()
		still := peer.gen == gen && peer.detached
		room.mu.Unlock()
		if still {
			s.finish(room, peer)
		}
	})
	room.mu.Unlock()
}

// finish removes a device from its room for good, and tells the others and the call history.
func (s *Server) finish(room *Room, peer *Peer) {
	empty := room.remove(peer)
	s.report(context.Background(), event{Type: "left", Call: room.ID, Org: room.Org, Person: peer.Person})
	if empty {
		s.closeRoom(room)
	}
}

func (s *Server) handle(room *Room, p *Peer, m *msg, raw []byte) {
	switch m.Type {
	case "offer", "answer", "ice", "key": // relayed untouched: the server cannot read or alter what the devices agree on
		if m.To == "" {
			return
		}
		m.From = p.ID
		out, _ := json.Marshal(m)
		room.sendTo(m.To, out)
	case "chat": // in-call chat: an opaque envelope for everyone in the room
		m.From = p.ID
		out, _ := json.Marshal(m)
		room.broadcastRaw(p.ID, out)
	case "state":
		room.mu.Lock()
		if m.Mic != nil {
			p.Mic = *m.Mic
		}
		if m.Cam != nil {
			p.Cam = *m.Cam
		}
		if m.Share != nil {
			p.Share = *m.Share
		}
		if m.NetPaused != nil {
			p.NetPaused = *m.NetPaused
		}
		if m.Hand != nil {
			p.Hand = *m.Hand
		}
		room.mu.Unlock()
		room.broadcast(p.ID, map[string]any{"type": "state", "peer": p})
	case "mod":
		s.moderate(room, p, m)
	case "ping":
		b, _ := json.Marshal(map[string]string{"type": "pong"})
		room.sendTo(p.ID, b)
	}
}

// moderate: hosts and moderators can mute a participant, turn their camera off, remove them, lock the room or end the call
// for everyone. The device obeys; the server only checks the sender is allowed to ask.
func (s *Server) moderate(room *Room, p *Peer, m *msg) {
	if p.Role != "host" && p.Role != "moderator" {
		b, _ := json.Marshal(map[string]string{"type": "error", "error": "only the host or a moderator can do that"})
		room.sendTo(p.ID, b)
		return
	}
	switch m.Action {
	case "mute", "cam_off", "remove":
		room.mu.Lock()
		t := room.peers[m.Target]
		room.mu.Unlock()
		if t == nil || (t.Role == "host" && p.Role != "host") {
			return
		}
		out, _ := json.Marshal(map[string]any{"type": "mod", "action": m.Action, "by": p.ID})
		room.sendTo(t.ID, out)
		if m.Action == "remove" {
			room.mu.Lock()
			t.final = true // removed on purpose: it must not be able to resume
			room.mu.Unlock()
			time.AfterFunc(500*time.Millisecond, func() { t.cancelNow() })
		}
	case "mute_all":
		room.broadcast(p.ID, map[string]any{"type": "mod", "action": "mute", "by": p.ID})
	case "lock", "unlock":
		room.mu.Lock()
		room.locked = m.Action == "lock"
		room.mu.Unlock()
		room.broadcast("", map[string]any{"type": "locked", "locked": m.Action == "lock"})
	case "end":
		if p.Role != "host" {
			return
		}
		room.broadcast("", map[string]any{"type": "ended", "by": p.ID})
		room.mu.Lock()
		room.ended = true
		ps := make([]*Peer, 0, len(room.peers))
		for _, x := range room.peers {
			x.final = true
			ps = append(ps, x)
		}
		room.mu.Unlock()
		time.AfterFunc(500*time.Millisecond, func() {
			for _, x := range ps {
				x.cancelNow()
			}
		})
	}
}

func (s *Server) room(c *calltoken.Claims) *Room {
	s.mu.Lock()
	defer s.mu.Unlock()
	r := s.rooms[c.Room]
	if r == nil || r.isEnded() {
		r = &Room{ID: c.Room, Org: c.Org, Kind: c.Kind, peers: map[string]*Peer{}}
		s.rooms[c.Room] = r
	}
	return r
}

func (s *Server) closeRoom(r *Room) {
	s.mu.Lock()
	if cur := s.rooms[r.ID]; cur == r {
		delete(s.rooms, r.ID)
	}
	s.mu.Unlock()
	s.report(context.Background(), event{Type: "ended", Call: r.ID, Org: r.Org})
}

func (r *Room) isEnded() bool { r.mu.Lock(); defer r.mu.Unlock(); return r.ended }

type roomError string

func (e roomError) Error() string { return string(e) }

func (r *Room) add(p *Peer, max int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if max <= 0 || max > maxPeersCap {
		max = maxPeersCap
	}
	if r.locked && p.Role == "participant" {
		return roomError("the host locked this room")
	}
	if len(r.peers) >= max {
		return roomError("this room is full")
	}
	r.peers[p.ID] = p
	return nil
}

func (r *Room) remove(p *Peer) (empty bool) {
	r.mu.Lock()
	delete(r.peers, p.ID)
	empty = len(r.peers) == 0
	r.mu.Unlock()
	if !empty {
		r.broadcast(p.ID, map[string]any{"type": "peer-left", "id": p.ID})
	}
	return empty
}

func (r *Room) sendTo(id string, b []byte) {
	r.mu.Lock()
	p := r.peers[id]
	r.mu.Unlock()
	if p == nil {
		return
	}
	select {
	case p.send <- b:
	default:
		p.cancelNow() // a stuck device is dropped instead of slowing the room
	}
}

func (r *Room) broadcast(except string, v any) {
	b, _ := json.Marshal(v)
	r.broadcastRaw(except, b)
}

func (r *Room) broadcastRaw(except string, b []byte) {
	r.mu.Lock()
	ps := make([]*Peer, 0, len(r.peers))
	for _, p := range r.peers {
		if p.ID != except {
			ps = append(ps, p)
		}
	}
	r.mu.Unlock()
	for _, p := range ps {
		select {
		case p.send <- b:
		default:
			p.cancelNow()
		}
	}
}

// cancelNow closes the device's current connection (the field is swapped when a device resumes, so it is read under no assumption
// that it is the one it was a moment ago).
func (p *Peer) cancelNow() {
	p.cmu.Lock()
	c := p.cancel
	p.cmu.Unlock()
	if c != nil {
		c()
	}
}

// Package callsvc is the calls microservice: it carries the signaling for voice and video rooms and the room controls. It
// never sees media (devices connect to each other, encrypted with DTLS-SRTP) and needs no database; it keeps rooms in
// memory and reports call events to Redis. Running it apart from the API means a big meeting cannot slow the rest of Hub.
package callsvc

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
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
	mu      sync.Mutex
	rooms   map[string]*Room
	peers   atomic.Int64
}

func New(secret []byte, rdb *redis.Client, origins []string) *Server {
	return &Server{Secret: secret, Rdb: rdb, Origins: origins, rooms: map[string]*Room{}}
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
	ID     string `json:"id"`
	Person string `json:"person"`
	Name   string `json:"name"`
	Role   string `json:"role"`
	Mic    bool   `json:"mic"`
	Cam    bool   `json:"cam"`
	Share  bool   `json:"sharing"`
	Hand   bool   `json:"hand"`
	send   chan []byte
	cancel context.CancelFunc
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
	return mux
}

type msg struct {
	Type   string          `json:"type"`
	To     string          `json:"to,omitempty"`
	From   string          `json:"from,omitempty"`
	Action string          `json:"action,omitempty"`
	Target string          `json:"target,omitempty"`
	Data   json.RawMessage `json:"data,omitempty"`
	Mic    *bool           `json:"mic,omitempty"`
	Cam    *bool           `json:"cam,omitempty"`
	Share  *bool           `json:"sharing,omitempty"`
	Hand   *bool           `json:"hand,omitempty"`
}

func (s *Server) ws(w http.ResponseWriter, r *http.Request) {
	claims, err := calltoken.Verify(s.Secret, r.URL.Query().Get("token"))
	if err != nil {
		http.Error(w, `{"error":"invalid or expired call pass"}`, http.StatusUnauthorized)
		return
	}
	// a pass works once
	if ok, _ := s.Rdb.SetNX(r.Context(), "callpass:"+claims.ID, 1, 3*time.Minute).Result(); !ok && claims.ID != "" {
		http.Error(w, `{"error":"this call pass was already used"}`, http.StatusUnauthorized)
		return
	}
	room := s.room(claims)
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	peer := &Peer{ID: uuid.NewString(), Person: claims.Person, Name: claims.Name, Role: claims.Role, Mic: true, Cam: claims.Kind == "video", send: make(chan []byte, sendBuffer), cancel: cancel}
	if err := room.add(peer, claims.MaxPeer); err != nil {
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
	defer func() {
		s.peers.Add(-1)
		empty := room.remove(peer)
		s.report(context.Background(), event{Type: "left", Call: room.ID, Org: room.Org, Person: peer.Person})
		if empty {
			s.closeRoom(room)
		}
		_ = conn.Close(websocket.StatusNormalClosure, "bye")
	}()

	room.mu.Lock()
	others := make([]*Peer, 0, len(room.peers))
	for _, p := range room.peers {
		if p.ID != peer.ID {
			others = append(others, p)
		}
	}
	room.mu.Unlock()
	welcome, _ := json.Marshal(map[string]any{"type": "welcome", "you": peer, "peers": others, "room": room.ID, "kind": room.Kind, "locked": room.locked})
	peer.send <- welcome
	room.broadcast(peer.ID, map[string]any{"type": "peer-joined", "peer": peer})

	go func() { // writer
		t := time.NewTicker(20 * time.Second)
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
				pc, c := context.WithTimeout(ctx, 8*time.Second)
				err := conn.Ping(pc)
				c()
				if err != nil {
					cancel()
					return
				}
			}
		}
	}()

	bucket, last := float64(msgsPerSec), time.Now()
	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			return
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
			time.AfterFunc(500*time.Millisecond, t.cancel)
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
			ps = append(ps, x)
		}
		room.mu.Unlock()
		time.AfterFunc(500*time.Millisecond, func() {
			for _, x := range ps {
				x.cancel()
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
		p.cancel() // a stuck device is dropped instead of slowing the room
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
			p.cancel()
		}
	}
}

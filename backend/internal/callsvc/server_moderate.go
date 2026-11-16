package callsvc

import (
	"context"
	"encoding/json"
	"time"

	"hub/internal/calltoken"
)

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

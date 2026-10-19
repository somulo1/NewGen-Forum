package realtime

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"
)

// Conn is one connected device.
type Conn struct {
	ID      string
	Account uuid.UUID
	Org     uuid.UUID
	send    chan []byte
	closed  atomic.Bool
	cancel  context.CancelFunc
}

// Hub keeps the in-memory routing table of one API node.
type Hub struct {
	mu      sync.RWMutex
	chans   map[string]map[*Conn]struct{}
	byConn  map[*Conn]map[string]struct{}
	dropped atomic.Int64
	conns   atomic.Int64
}

func NewHub() *Hub {
	return &Hub{chans: map[string]map[*Conn]struct{}{}, byConn: map[*Conn]map[string]struct{}{}}
}

func (h *Hub) Connections() int64 { return h.conns.Load() }

// Subscribe attaches a connection to a channel. The caller must already have checked the account may read it.
func (h *Hub) Subscribe(c *Conn, channels ...string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, ch := range channels {
		if h.chans[ch] == nil {
			h.chans[ch] = map[*Conn]struct{}{}
		}
		h.chans[ch][c] = struct{}{}
		if h.byConn[c] == nil {
			h.byConn[c] = map[string]struct{}{}
		}
		h.byConn[c][ch] = struct{}{}
	}
}

func (h *Hub) Unsubscribe(c *Conn, channels ...string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, ch := range channels {
		delete(h.chans[ch], c)
		if len(h.chans[ch]) == 0 {
			delete(h.chans, ch)
		}
		delete(h.byConn[c], ch)
	}
}

func (h *Hub) remove(c *Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.byConn[c] {
		delete(h.chans[ch], c)
		if len(h.chans[ch]) == 0 {
			delete(h.chans, ch)
		}
	}
	delete(h.byConn, c)
}

// Deliver routes an event to local subscribers. A slow client never blocks others: its queue is bounded and it is
// disconnected (to resync) if it falls too far behind.
func (h *Hub) Deliver(e Event) {
	raw, _ := json.Marshal(e)
	h.mu.RLock()
	targets := make([]*Conn, 0, len(h.chans[e.Channel]))
	for c := range h.chans[e.Channel] {
		targets = append(targets, c)
	}
	h.mu.RUnlock()
	for _, c := range targets {
		select {
		case c.send <- raw:
		default:
			h.dropped.Add(1)
			c.cancel()
		}
	}
	if e.Type == "session.ended" { // the account was signed out or removed: close its sockets on this node
		for _, c := range targets {
			if c.Account.String() == strings.TrimPrefix(e.Channel, "acct:") {
				go func(c *Conn) { time.Sleep(300 * time.Millisecond); c.cancel() }(c)
			}
		}
	}
}

// Count is how many channels a connection is attached to.
func (h *Hub) Count(c *Conn) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.byConn[c])
}

// Subscribed reports whether the connection is attached to the channel.
func (h *Hub) Subscribed(c *Conn, ch string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.byConn[c][ch]
	return ok
}

// Kick disconnects every connection of an account from one channel's audience (used when someone is removed from a conversation).
func (h *Hub) Detach(account uuid.UUID, channel string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.chans[channel] {
		if c.Account == account {
			delete(h.chans[channel], c)
			delete(h.byConn[c], channel)
		}
	}
}

// Serve upgrades the request and pumps events until the client leaves. onMessage handles client frames (typing, ping, ...).
func (h *Hub) Serve(ctx context.Context, w ws, acct, org uuid.UUID, channels []string, onMessage func(*Conn, []byte)) error {
	conn, err := w.Accept()
	if err != nil {
		return err
	}
	conn.SetReadLimit(64 << 10)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	c := &Conn{ID: uuid.NewString(), Account: acct, Org: org, send: make(chan []byte, 256), cancel: cancel}
	h.Subscribe(c, channels...)
	h.conns.Add(1)
	defer func() {
		h.conns.Add(-1)
		h.remove(c)
		c.closed.Store(true)
		_ = conn.Close(websocket.StatusNormalClosure, "bye")
	}()

	go func() { // writer
		t := time.NewTicker(25 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case b := <-c.send:
				wctx, wc := context.WithTimeout(ctx, 10*time.Second)
				err := conn.Write(wctx, websocket.MessageText, b)
				wc()
				if err != nil {
					cancel()
					return
				}
			case <-t.C:
				pctx, pc := context.WithTimeout(ctx, 10*time.Second)
				err := conn.Ping(pctx)
				pc()
				if err != nil {
					cancel()
					return
				}
			}
		}
	}()
	for { // reader
		typ, data, err := conn.Read(ctx)
		if err != nil {
			return nil
		}
		if typ == websocket.MessageText && onMessage != nil {
			onMessage(c, data)
		}
	}
}

// Send queues a frame to one connection.
func (c *Conn) Send(v any) {
	b, _ := json.Marshal(v)
	select {
	case c.send <- b:
	default:
		slog.Warn("client queue full", "conn", c.ID)
		c.cancel()
	}
}

// ws abstracts the HTTP upgrade so the hub can be tested without a network.
type ws interface {
	Accept() (*websocket.Conn, error)
}

// Upgrade adapts an HTTP request to the hub. Origins limits which websites may open a socket (cross-site hijacking guard).
type Upgrade struct {
	W       http.ResponseWriter
	R       *http.Request
	Origins []string
}

func (u Upgrade) Accept() (*websocket.Conn, error) {
	patterns := make([]string, 0, len(u.Origins))
	for _, o := range u.Origins {
		o = strings.TrimPrefix(strings.TrimPrefix(o, "https://"), "http://")
		patterns = append(patterns, o)
	}
	return websocket.Accept(u.W, u.R, &websocket.AcceptOptions{OriginPatterns: patterns})
}

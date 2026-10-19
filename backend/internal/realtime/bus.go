// Package realtime delivers changes to connected clients as they happen.
//
// Every state change is published once to a Redis channel; every API node holds ONE subscription for all channels and
// routes events to its own WebSocket connections in memory. Adding nodes adds capacity; nobody polls the database.
package realtime

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Event is the unit of change. For end-to-end encrypted content, Data carries only ciphertext envelopes.
type Event struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`    // e.g. "message.new", "wall.post", "poll.vote", "ticket.updated", "presence"
	Channel string          `json:"channel"` // org:<id> | conv:<id> | acct:<id>
	Org     string          `json:"org"`
	Actor   string          `json:"actor,omitempty"`
	TS      time.Time       `json:"ts"`
	Data    json.RawMessage `json:"data,omitempty"`
}

const prefix = "hub:ev:"

// Bus publishes and receives events across all nodes.
type Bus struct{ rdb *redis.Client }

func NewBus(rdb *redis.Client) *Bus { return &Bus{rdb: rdb} }

func OrgChannel(org uuid.UUID) string { return "org:" + org.String() }
func ConvChannel(c uuid.UUID) string  { return "conv:" + c.String() }
func AcctChannel(a uuid.UUID) string  { return "acct:" + a.String() }

// Publish sends an event. Failures are returned so callers can decide whether the change still counts (it normally does:
// the database is the source of truth and clients re-sync on reconnect).
func (b *Bus) Publish(ctx context.Context, e Event) error {
	if e.ID == "" {
		e.ID = uuid.NewString()
	}
	if e.TS.IsZero() {
		e.TS = time.Now().UTC()
	}
	raw, err := json.Marshal(e)
	if err != nil {
		return err
	}
	return b.rdb.Publish(ctx, prefix+e.Channel, raw).Err()
}

// Run subscribes to every Hub channel and calls fn for each event until ctx ends. It reconnects by itself.
func (b *Bus) Run(ctx context.Context, fn func(Event)) {
	for ctx.Err() == nil {
		ps := b.rdb.PSubscribe(ctx, prefix+"*")
		ch := ps.Channel(redis.WithChannelSize(4096))
	loop:
		for {
			select {
			case <-ctx.Done():
				_ = ps.Close()
				return
			case m, ok := <-ch:
				if !ok {
					break loop
				}
				var e Event
				if json.Unmarshal([]byte(m.Payload), &e) == nil && strings.HasPrefix(m.Channel, prefix) {
					e.Channel = strings.TrimPrefix(m.Channel, prefix)
					fn(e)
				}
			}
		}
		_ = ps.Close()
		select {
		case <-ctx.Done():
		case <-time.After(time.Second):
		}
	}
}

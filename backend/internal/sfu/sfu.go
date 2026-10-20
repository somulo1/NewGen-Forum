// Package sfu is the API's link to the media servers (LiveKit). Devices send their audio and video to a media server, which
// forwards it to everyone else, so a room is no longer limited by how many streams one device can carry. This package only
// speaks the control side: it mints the pass a device uses to join, runs the room controls (mute, remove, bring on stage, end)
// and checks the signed notices a media server sends when people join and leave. Media never passes through the API.
package sfu

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"hub/internal/config"
)

type Client struct {
	Nodes []config.SFUNode
	http  *http.Client
}

func New(nodes []config.SFUNode) *Client {
	return &Client{Nodes: nodes, http: &http.Client{Timeout: 8 * time.Second}}
}

// On reports whether any media server is configured.
func (c *Client) On() bool { return c != nil && len(c.Nodes) > 0 }

// Pick pins a room to one media server (or region) by rendezvous hashing, so every API node picks the same one and
// adding a server moves only a small share of rooms.
func (c *Client) Pick(room string) config.SFUNode {
	best, bestScore := c.Nodes[0], uint64(0)
	for i, n := range c.Nodes {
		h := fnv.New64a()
		h.Write([]byte(n.URL + "|" + room))
		if s := h.Sum64(); i == 0 || s > bestScore {
			best, bestScore = n, s
		}
	}
	return best
}

// WS is the address devices connect to.
func WS(n config.SFUNode) string {
	return strings.Replace(strings.Replace(n.URL, "http://", "ws://", 1), "https://", "wss://", 1)
}

// Grant is what a pass allows.
type Grant struct {
	Room      string
	Publish   bool
	Sources   []string // microphone | camera | screen_share | screen_share_audio
	Admin     bool     // room controls (used by the API itself)
	Identity  string
	Name      string
	Metadata  string
	Hidden    bool
	ExpiresIn time.Duration
}

type videoGrant struct {
	RoomJoin             bool     `json:"roomJoin,omitempty"`
	RoomAdmin            bool     `json:"roomAdmin,omitempty"`
	RoomCreate           bool     `json:"roomCreate,omitempty"`
	RoomList             bool     `json:"roomList,omitempty"`
	Room                 string   `json:"room,omitempty"`
	CanPublish           *bool    `json:"canPublish,omitempty"`
	CanSubscribe         *bool    `json:"canSubscribe,omitempty"`
	CanPublishData       *bool    `json:"canPublishData,omitempty"`
	CanPublishSources    []string `json:"canPublishSources,omitempty"`
	CanUpdateOwnMetadata *bool    `json:"canUpdateOwnMetadata,omitempty"`
}

type claims struct {
	jwt.RegisteredClaims
	Name     string      `json:"name,omitempty"`
	Metadata string      `json:"metadata,omitempty"`
	SHA256   string      `json:"sha256,omitempty"`
	Video    *videoGrant `json:"video,omitempty"`
}

func ptr[T any](v T) *T { return &v }

// Token mints a join pass for one device.
func Token(n config.SFUNode, g Grant) (string, error) {
	ttl := g.ExpiresIn
	if ttl == 0 {
		ttl = 2 * time.Minute
	}
	now := time.Now()
	v := &videoGrant{RoomJoin: true, Room: g.Room, CanSubscribe: ptr(true), CanPublish: ptr(g.Publish), CanPublishData: ptr(true), CanUpdateOwnMetadata: ptr(true)}
	if g.Publish {
		v.CanPublishSources = g.Sources
	}
	c := claims{RegisteredClaims: jwt.RegisteredClaims{Issuer: n.Key, Subject: g.Identity, NotBefore: jwt.NewNumericDate(now.Add(-10 * time.Second)), ExpiresAt: jwt.NewNumericDate(now.Add(ttl))}, Name: g.Name, Metadata: g.Metadata, Video: v}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString([]byte(n.Secret))
}

func adminToken(n config.SFUNode, room string) (string, error) {
	now := time.Now()
	c := claims{RegisteredClaims: jwt.RegisteredClaims{Issuer: n.Key, Subject: "hub-api", NotBefore: jwt.NewNumericDate(now.Add(-10 * time.Second)), ExpiresAt: jwt.NewNumericDate(now.Add(time.Minute))},
		Video: &videoGrant{RoomAdmin: true, RoomCreate: true, RoomList: true, Room: room}}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString([]byte(n.Secret))
}

// ErrFull is returned by the media server when a room is at its limit.
var ErrFull = errors.New("sfu: room full")

func (c *Client) rpc(ctx context.Context, n config.SFUNode, method, room string, in, out any) error {
	tok, err := adminToken(n, room)
	if err != nil {
		return err
	}
	b, _ := json.Marshal(in)
	req, _ := http.NewRequestWithContext(ctx, "POST", n.URL+"/twirp/livekit.RoomService/"+method, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tok)
	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(res.Body, 4<<20))
	if res.StatusCode >= 300 {
		return fmt.Errorf("media server %s: %d %s", method, res.StatusCode, strings.TrimSpace(string(body)))
	}
	if out != nil {
		return json.Unmarshal(body, out)
	}
	return nil
}

// CreateRoom opens (or confirms) a room with a hard cap on people. emptyFor is how long a room with nobody in it stays open.
func (c *Client) CreateRoom(ctx context.Context, room string, max int, emptyFor time.Duration) error {
	return c.rpc(ctx, c.Pick(room), "CreateRoom", room, map[string]any{"name": room, "emptyTimeout": int(emptyFor.Seconds()), "departureTimeout": 20, "maxParticipants": max}, nil)
}

func (c *Client) DeleteRoom(ctx context.Context, room string) error {
	return c.rpc(ctx, c.Pick(room), "DeleteRoom", room, map[string]any{"room": room}, nil)
}

func (c *Client) RemoveParticipant(ctx context.Context, room, identity string) error {
	return c.rpc(ctx, c.Pick(room), "RemoveParticipant", room, map[string]any{"room": room, "identity": identity}, nil)
}

// Track and Participant are the few fields of the media server's view of a room that Hub uses.
type Track struct {
	Sid    string `json:"sid"`
	Type   string `json:"type"`
	Source string `json:"source"`
	Muted  bool   `json:"muted"`
}

type Participant struct {
	Identity   string  `json:"identity"`
	Name       string  `json:"name"`
	State      string  `json:"state"`
	Tracks     []Track `json:"tracks"`
	Permission struct {
		CanPublish bool `json:"canPublish"`
	} `json:"permission"`
}

func (c *Client) List(ctx context.Context, room string) ([]Participant, error) {
	var out struct {
		Participants []Participant `json:"participants"`
	}
	err := c.rpc(ctx, c.Pick(room), "ListParticipants", room, map[string]any{"room": room}, &out)
	return out.Participants, err
}

// SetStage lets a listener speak (or sends a speaker back to listening).
func (c *Client) SetStage(ctx context.Context, room, identity string, on bool, sources []string) error {
	perm := map[string]any{"canSubscribe": true, "canPublish": on, "canPublishData": true}
	if on {
		perm["canPublishSources"] = sources
	}
	return c.rpc(ctx, c.Pick(room), "UpdateParticipant", room, map[string]any{"room": room, "identity": identity, "permission": perm}, nil)
}

// Mute silences someone's microphone (audio=true) or camera. A person can turn their own back on; a moderator can only turn it off.
func (c *Client) Mute(ctx context.Context, room, identity string, audio bool) error {
	ps, err := c.List(ctx, room)
	if err != nil {
		return err
	}
	for _, p := range ps {
		if p.Identity != identity {
			continue
		}
		for _, t := range p.Tracks {
			src := strings.ToUpper(t.Source)
			if (audio && src == "MICROPHONE") || (!audio && src == "CAMERA") {
				if err := c.rpc(ctx, c.Pick(room), "MutePublishedTrack", room, map[string]any{"room": room, "identity": identity, "trackSid": t.Sid, "muted": true}, nil); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// MuteAll silences every microphone except those of the people named in keep.
func (c *Client) MuteAll(ctx context.Context, room string, keep map[string]bool) error {
	ps, err := c.List(ctx, room)
	if err != nil {
		return err
	}
	for _, p := range ps {
		if keep[p.Identity] {
			continue
		}
		for _, t := range p.Tracks {
			if strings.EqualFold(t.Source, "MICROPHONE") && !t.Muted {
				_ = c.rpc(ctx, c.Pick(room), "MutePublishedTrack", room, map[string]any{"room": room, "identity": p.Identity, "trackSid": t.Sid, "muted": true}, nil)
			}
		}
	}
	return nil
}

// Event is a signed notice from a media server.
type Event struct {
	Event string `json:"event"` // room_started | room_finished | participant_joined | participant_left
	Room  struct {
		Name string `json:"name"`
	} `json:"room"`
	Participant struct {
		Identity string `json:"identity"`
		Metadata string `json:"metadata"`
	} `json:"participant"`
}

// Verify checks that a notice really came from one of our media servers and was not altered on the way.
func (c *Client) Verify(authorization string, body []byte) (*Event, error) {
	tok := strings.TrimPrefix(authorization, "Bearer ")
	var cl claims
	_, err := jwt.ParseWithClaims(tok, &cl, func(t *jwt.Token) (any, error) {
		iss, _ := t.Claims.GetIssuer()
		for _, n := range c.Nodes {
			if n.Key == iss {
				return []byte(n.Secret), nil
			}
		}
		return nil, errors.New("unknown media server")
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired(), jwt.WithLeeway(30*time.Second))
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(body)
	if cl.SHA256 != base64.StdEncoding.EncodeToString(sum[:]) {
		return nil, errors.New("notice does not match its signature")
	}
	var e Event
	if err := json.Unmarshal(body, &e); err != nil {
		return nil, err
	}
	return &e, nil
}

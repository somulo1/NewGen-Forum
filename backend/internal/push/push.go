// Package push delivers notifications to the Hub phone app when it is in the background or closed, through Firebase Cloud
// Messaging (Android) and Apple's push service (iOS). It listens to the same live events the open app receives: a bell
// notification, an incoming call, a new chat message. Chat messages are end-to-end encrypted, so a push never carries what was
// written, and the organization's "details in emails" setting decides whether it may name the sender and the chat.
package push

import (
	"context"
	"encoding/base64"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"

	"hub/internal/app"
	"hub/internal/realtime"
)

const (
	channelID  = "hub" // the Android notification channel the app creates
	accent     = "#B8F36B"
	queueSize  = 2000
	workers    = 4
	sendWithin = 15 * time.Second
)

// Note is one notification as the phone shows it.
type Note struct {
	Title, Body, Tag string
	Urgent           bool              // an incoming call: delivered at once or not at all
	Data             map[string]string // where a tap leads (ids only)
}

type job struct {
	org      uuid.UUID
	accounts []uuid.UUID
	note     Note
}

type sender interface {
	send(ctx context.Context, token string, n Note) error
}

// Service turns live events into pushes.
type Service struct {
	a       *app.App
	android sender
	ios     sender
	events  chan realtime.Event
	details *detailCache
}

// Start reads the credentials from the environment, hooks into the live events and starts the senders. With no credentials
// the phones still register; nothing is sent until they are set.
func Start(ctx context.Context, a *app.App) *Service {
	client := &http.Client{Timeout: 10 * time.Second, Transport: http.DefaultTransport.(*http.Transport).Clone()}
	s := &Service{a: a, events: make(chan realtime.Event, queueSize), details: newDetailCache()}
	if creds := secret("HUB_FCM_CREDENTIALS"); len(creds) > 0 {
		if f, err := newFCM(creds, client); err != nil {
			slog.Error("push: android off", "err", err)
		} else {
			s.android = f
		}
	}
	if key := secret("HUB_APNS_KEY"); len(key) > 0 {
		topic := strings.TrimSpace(os.Getenv("HUB_APNS_TOPIC"))
		if topic == "" {
			topic = "com.newgen.hub"
		}
		if c, err := newAPNs(key, strings.TrimSpace(os.Getenv("HUB_APNS_KEY_ID")), strings.TrimSpace(os.Getenv("HUB_APNS_TEAM_ID")), topic,
			strings.EqualFold(os.Getenv("HUB_APNS_SANDBOX"), "true"), client); err != nil {
			slog.Error("push: ios off", "err", err)
		} else {
			s.ios = c
		}
	}
	slog.Info("push notifications", "android", s.android != nil, "ios", s.ios != nil)
	if s.android == nil && s.ios == nil {
		return s
	}
	a.OnPublish = s.observe
	for i := 0; i < workers; i++ {
		go s.run(ctx)
	}
	return s
}

// secret reads a credential given as its text (JSON or PEM), as base64 of it, or as the path of a file holding it.
func secret(name string) []byte {
	v := strings.TrimSpace(os.Getenv(name))
	switch {
	case v == "":
		return nil
	case strings.HasPrefix(v, "{") || strings.HasPrefix(v, "-----BEGIN"):
		return []byte(v)
	}
	if b, err := base64.StdEncoding.DecodeString(v); err == nil {
		return b
	}
	if b, err := os.ReadFile(v); err == nil {
		return b
	}
	slog.Error("push: credential unreadable", "env", name)
	return nil
}

// observe is called for every live event; it must never slow the caller, so a full queue drops the push (the app still shows
// everything when opened).
func (s *Service) observe(_ context.Context, e realtime.Event) {
	if !wanted(e) {
		return
	}
	select {
	case s.events <- e:
	default:
		slog.Warn("push: queue full, notification dropped", "type", e.Type)
	}
}

func (s *Service) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case e := <-s.events:
			jc, cancel := context.WithTimeout(ctx, sendWithin)
			if j, ok := s.compose(jc, e); ok {
				s.deliver(jc, j)
			}
			cancel()
		}
	}
}

func (s *Service) deliver(ctx context.Context, j job) {
	if len(j.accounts) == 0 {
		return
	}
	devices, err := devicesOf(ctx, s.a, j.org, j.accounts)
	if err != nil {
		slog.Warn("push: devices not read", "err", err)
		return
	}
	for _, d := range devices {
		via := s.android
		if d.platform == "ios" {
			via = s.ios
		}
		if via == nil {
			continue
		}
		if err := via.send(ctx, d.token, j.note); err == errGone {
			forget(ctx, s.a, d.hash)
		} else if err != nil {
			slog.Warn("push: not delivered", "platform", d.platform, "err", err)
		}
	}
}

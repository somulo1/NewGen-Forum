// Package mail sends system email. Production plugs an SMTP/SES sender into the same interface.
package mail

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type Message struct {
	To, Subject, Body string
	At                time.Time
	HTML              string
}

type Mailer interface {
	Send(ctx context.Context, to, subject, body string) error
}

// Dev keeps the last messages in memory and logs them, standing in for an inbox while there is no mail server.
// The body of a dev message contains the one-time codes, so Dev must never be used when HUB_ENV=prod.
type Dev struct {
	mu    sync.Mutex
	box   []Message
	Quiet bool
	Cap   int // how many messages to keep (default 200)
}

func (d *Dev) Send(_ context.Context, to, subject, body string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.box = append(d.box, Message{To: to, Subject: subject, Body: body, At: time.Now()})
	max := d.Cap
	if max == 0 {
		max = 200
	}
	if len(d.box) > max {
		d.box = d.box[len(d.box)-max:]
	}
	if !d.Quiet {
		slog.Info("dev mail", "to", to, "subject", subject)
	}
	return nil
}

// SendHTML keeps the message with its HTML version, so a test (or the developer) can look at it.
func (d *Dev) SendHTML(ctx context.Context, to, subject, text, html string) error {
	if err := d.Send(ctx, to, subject, text); err != nil {
		return err
	}
	d.mu.Lock()
	d.box[len(d.box)-1].HTML = html
	d.mu.Unlock()
	return nil
}

// To returns messages for one address, newest first.
func (d *Dev) To(addr string) []Message {
	d.mu.Lock()
	defer d.mu.Unlock()
	var out []Message
	for i := len(d.box) - 1; i >= 0; i-- {
		if d.box[i].To == addr {
			out = append(out, d.box[i])
		}
	}
	return out
}

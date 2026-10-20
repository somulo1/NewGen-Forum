package mail

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"strings"
	"time"
)

// HTMLMailer is a Mailer that can also send a message with an HTML version.
type HTMLMailer interface {
	Mailer
	SendHTML(ctx context.Context, to, subject, text, html string) error
}

// SendCard sends a card by the best way the mailer offers.
func SendCard(ctx context.Context, m Mailer, to, subject string, c Card) error {
	if h, ok := m.(HTMLMailer); ok {
		return h.SendHTML(ctx, to, subject, c.Text(), c.HTML())
	}
	return m.Send(ctx, to, subject, c.Text())
}

// SMTP sends through a mail server. It is chosen when HUB_SMTP_HOST is set:
//
//	HUB_SMTP_HOST, HUB_SMTP_PORT (default 587), HUB_SMTP_USER, HUB_SMTP_PASS,
//	HUB_SMTP_FROM (the sending address), HUB_SMTP_FROM_NAME (default "Hub"),
//	HUB_SMTP_SECURITY: starttls (default), tls (implicit TLS, usually port 465) or none.
type SMTP struct {
	Host, Port, User, Pass, From, FromName, Security string
}

// FromEnv returns the SMTP sender described by the environment, or nil when none is set up.
func FromEnv() *SMTP {
	host := strings.TrimSpace(os.Getenv("HUB_SMTP_HOST"))
	if host == "" {
		return nil
	}
	get := func(k, d string) string {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			return v
		}
		return d
	}
	s := &SMTP{Host: host, Port: get("HUB_SMTP_PORT", "587"), User: os.Getenv("HUB_SMTP_USER"), Pass: os.Getenv("HUB_SMTP_PASS"),
		FromName: get("HUB_SMTP_FROM_NAME", "Hub"), Security: strings.ToLower(get("HUB_SMTP_SECURITY", "starttls"))}
	s.From = get("HUB_SMTP_FROM", s.User)
	return s
}

func (s *SMTP) Send(ctx context.Context, to, subject, body string) error {
	return s.deliver(ctx, to, subject, body, "")
}

func (s *SMTP) SendHTML(ctx context.Context, to, subject, text, html string) error {
	return s.deliver(ctx, to, subject, text, html)
}

func (s *SMTP) deliver(ctx context.Context, to, subject, text, htmlBody string) error {
	if strings.ContainsAny(to+subject, "\r\n") {
		return fmt.Errorf("mail: bad header value")
	}
	addr := net.JoinHostPort(s.Host, s.Port)
	d := net.Dialer{Timeout: 15 * time.Second}
	var conn net.Conn
	var err error
	if s.Security == "tls" {
		conn, err = tls.DialWithDialer(&d, "tcp", addr, &tls.Config{ServerName: s.Host, MinVersion: tls.VersionTLS12})
	} else {
		conn, err = d.DialContext(ctx, "tcp", addr)
	}
	if err != nil {
		return err
	}
	_ = conn.SetDeadline(time.Now().Add(45 * time.Second))
	c, err := smtp.NewClient(conn, s.Host)
	if err != nil {
		conn.Close()
		return err
	}
	defer c.Close()
	if s.Security == "starttls" {
		if err := c.StartTLS(&tls.Config{ServerName: s.Host, MinVersion: tls.VersionTLS12}); err != nil {
			return err
		}
	}
	if s.User != "" {
		if err := c.Auth(smtp.PlainAuth("", s.User, s.Pass, s.Host)); err != nil {
			return err
		}
	}
	if err := c.Mail(s.From); err != nil {
		return err
	}
	if err := c.Rcpt(to); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write([]byte(s.message(to, subject, text, htmlBody))); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return c.Quit()
}

func (s *SMTP) message(to, subject, text, htmlBody string) string {
	from := (&mail.Address{Name: s.FromName, Address: s.From}).String()
	id := make([]byte, 12)
	_, _ = rand.Read(id)
	dom := "hub.local"
	if i := strings.LastIndex(s.From, "@"); i >= 0 {
		dom = s.From[i+1:]
	}
	h := "From: " + from + "\r\nTo: " + to + "\r\nSubject: " + mime.QEncoding.Encode("utf-8", subject) + "\r\nDate: " + time.Now().Format(time.RFC1123Z) +
		"\r\nMessage-ID: <" + hex.EncodeToString(id) + "@" + dom + ">\r\nMIME-Version: 1.0\r\nAuto-Submitted: auto-generated\r\n"
	enc := func(v string) string {
		b := base64.StdEncoding.EncodeToString([]byte(v))
		var out strings.Builder
		for len(b) > 76 {
			out.WriteString(b[:76] + "\r\n")
			b = b[76:]
		}
		return out.String() + b + "\r\n"
	}
	if htmlBody == "" {
		return h + "Content-Type: text/plain; charset=utf-8\r\nContent-Transfer-Encoding: base64\r\n\r\n" + enc(text)
	}
	bd := "hub" + hex.EncodeToString(id)
	return h + "Content-Type: multipart/alternative; boundary=" + bd + "\r\n\r\n" +
		"--" + bd + "\r\nContent-Type: text/plain; charset=utf-8\r\nContent-Transfer-Encoding: base64\r\n\r\n" + enc(text) +
		"--" + bd + "\r\nContent-Type: text/html; charset=utf-8\r\nContent-Transfer-Encoding: base64\r\n\r\n" + enc(htmlBody) +
		"--" + bd + "--\r\n"
}

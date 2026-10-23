package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

// API sends through an email service's web API (HTTPS, port 443). Some hosts, Render's free instances among them, block every
// outgoing SMTP port, so SMTP cannot be used there. It is chosen when HUB_MAIL_API is set:
//
//	HUB_MAIL_API       brevo | resend
//	HUB_MAIL_API_KEY   the service's API key
//	HUB_MAIL_FROM      the sending address (must be one the service has verified for you)
//	HUB_MAIL_FROM_NAME the name shown (default "Hub")
type API struct {
	Provider, Key, From, FromName string
	Client                        *http.Client
}

// APIFromEnv returns the web-API sender described by the environment, or nil when none is set up.
func APIFromEnv() *API {
	p := strings.ToLower(strings.TrimSpace(os.Getenv("HUB_MAIL_API")))
	if p == "" {
		return nil
	}
	name := strings.TrimSpace(os.Getenv("HUB_MAIL_FROM_NAME"))
	if name == "" {
		name = "Hub"
	}
	return &API{Provider: p, Key: strings.TrimSpace(os.Getenv("HUB_MAIL_API_KEY")), From: strings.TrimSpace(os.Getenv("HUB_MAIL_FROM")), FromName: name,
		Client: &http.Client{Timeout: 20 * time.Second}}
}

func (a *API) Send(ctx context.Context, to, subject, body string) error {
	return a.deliver(ctx, to, subject, body, "")
}

func (a *API) SendHTML(ctx context.Context, to, subject, text, html string) error {
	return a.deliver(ctx, to, subject, text, html)
}

func (a *API) deliver(ctx context.Context, to, subject, text, html string) error {
	if strings.ContainsAny(to+subject, "\r\n") {
		return fmt.Errorf("mail: bad header value")
	}
	if a.Key == "" || a.From == "" {
		return fmt.Errorf("mail: HUB_MAIL_API_KEY and HUB_MAIL_FROM are required")
	}
	var url string
	var payload map[string]any
	headers := map[string]string{"Content-Type": "application/json", "Accept": "application/json"}
	switch a.Provider {
	case "brevo":
		url = "https://api.brevo.com/v3/smtp/email"
		headers["api-key"] = a.Key
		payload = map[string]any{"sender": map[string]string{"name": a.FromName, "email": a.From}, "to": []map[string]string{{"email": to}}, "subject": subject, "textContent": text}
		if html != "" {
			payload["htmlContent"] = html
		}
	case "resend":
		url = "https://api.resend.com/emails"
		headers["Authorization"] = "Bearer " + a.Key
		payload = map[string]any{"from": a.FromName + " <" + a.From + ">", "to": []string{to}, "subject": subject, "text": text}
		if html != "" {
			payload["html"] = html
		}
	default:
		return fmt.Errorf("mail: unknown HUB_MAIL_API %q (use brevo or resend)", a.Provider)
	}
	b, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res, err := a.Client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode/100 != 2 {
		msg, _ := io.ReadAll(io.LimitReader(res.Body, 400))
		return fmt.Errorf("mail: %s answered %d: %s", a.Provider, res.StatusCode, strings.TrimSpace(string(msg)))
	}
	return nil
}

// Logged wraps a mailer so that a message that could not be sent says so in the log (most callers send in the background and
// do not look at the result, so a refused or unreachable mail service would otherwise go unnoticed).
type Logged struct{ M Mailer }

func (l Logged) Send(ctx context.Context, to, subject, body string) error {
	err := l.M.Send(ctx, to, subject, body)
	l.note(to, subject, err)
	return err
}

func (l Logged) SendHTML(ctx context.Context, to, subject, text, html string) error {
	var err error
	if h, ok := l.M.(HTMLMailer); ok {
		err = h.SendHTML(ctx, to, subject, text, html)
	} else {
		err = l.M.Send(ctx, to, subject, text)
	}
	l.note(to, subject, err)
	return err
}

func (l Logged) note(to, subject string, err error) {
	at := "***"
	if i := strings.Index(to, "@"); i > 1 {
		at = to[:1] + "***" + to[i:]
	}
	if err != nil {
		slog.Error("email not sent", "to", at, "subject", subject, "err", err)
		return
	}
	slog.Info("email sent", "to", at, "subject", subject)
}

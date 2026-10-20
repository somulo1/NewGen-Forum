package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"hub/internal/httpx"
	"hub/internal/mail"
	"hub/internal/tenant"
)

// Middleware requires a valid access token and puts the caller in the request context.
func (s *Service) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			httpx.Wrap(func(http.ResponseWriter, *http.Request) error { return httpx.ErrUnauthorized })(w, r)
			return
		}
		c, err := s.Verify(r.Context(), strings.TrimPrefix(h, "Bearer "))
		if err != nil {
			httpx.Wrap(func(http.ResponseWriter, *http.Request) error { return err })(w, r)
			return
		}
		if c.MC && r.URL.Path != "/v1/auth/password" && r.URL.Path != "/v1/auth/logout" {
			httpx.Wrap(func(http.ResponseWriter, *http.Request) error {
				return httpx.Err(403, "password_change_required", "Choose a new password to continue.")
			})(w, r)
			return
		}
		// inactive past the organization's limit: the session ends now (signing out still works)
		if c.Org != "" && r.URL.Path != "/v1/auth/logout" {
			org, _ := uuid.Parse(c.Org)
			sid, _ := uuid.Parse(c.Sid)
			if s.idleExpired(r.Context(), org, sid) {
				s.endIdle(r.Context(), sid)
				httpx.Wrap(func(http.ResponseWriter, *http.Request) error { return errIdle })(w, r)
				return
			}
		}
		p := &tenant.Principal{Name: c.Name, IP: httpx.ClientIP(r)}
		p.Account, _ = uuid.Parse(c.Subject)
		p.Session, _ = uuid.Parse(c.Sid)
		if c.Org != "" {
			p.Org, _ = uuid.Parse(c.Org)
			p.Person, _ = uuid.Parse(c.Pid)
		}
		next.ServeHTTP(w, r.WithContext(tenant.With(r.Context(), p)))
	})
}

// Public registers the endpoints that need no sign-in.
func (s *Service) Public(r chi.Router) {
	r.Post("/auth/signup", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		var in struct {
			Name, Email, Password string
			Quiz                  []QuizIn `json:"quiz"`
		}
		if err := httpx.Decode(r, &in, 16<<10); err != nil {
			return err
		}
		id, err := s.Signup(r.Context(), in.Name, in.Email, in.Password, in.Quiz, httpx.ClientIP(r))
		if err != nil {
			return err
		}
		httpx.JSON(w, 201, map[string]any{"account": id})
		return nil
	}))
	r.Post("/auth/login", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		var in struct {
			Email    string `json:"email"`
			Password string `json:"password"`
			TOTP     string `json:"totp"`
			Org      string `json:"org"`
		}
		if err := httpx.Decode(r, &in, 8<<10); err != nil {
			return err
		}
		t, err := s.Login(r.Context(), LoginIn{in.Email, in.Password, in.TOTP, in.Org, httpx.ClientIP(r), r.UserAgent()})
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, t)
		return nil
	}))
	r.Post("/auth/refresh", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		var in struct {
			Refresh string `json:"refresh_token"`
			Org     string `json:"org"`
		}
		if err := httpx.Decode(r, &in, 4<<10); err != nil {
			return err
		}
		t, err := s.Refresh(r.Context(), in.Refresh, in.Org, httpx.ClientIP(r), r.UserAgent())
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, t)
		return nil
	}))

	// password recovery (see reset.go)
	r.Post("/auth/reset/start", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		var in struct{ Email string }
		if err := httpx.Decode(r, &in, 2<<10); err != nil {
			return err
		}
		id, err := s.ResetStart(r.Context(), in.Email, httpx.ClientIP(r))
		if err != nil {
			return err
		}
		httpx.JSON(w, 202, map[string]any{"reset_id": id})
		return nil
	}))
	r.Post("/auth/reset/code", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		var in struct {
			ResetID uuid.UUID `json:"reset_id"`
			Code    string    `json:"code"`
		}
		if err := httpx.Decode(r, &in, 2<<10); err != nil {
			return err
		}
		qs, err := s.ResetVerifyCode(r.Context(), in.ResetID, in.Code)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"questions": qs})
		return nil
	}))
	r.Post("/auth/reset/quiz", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		var in struct {
			ResetID uuid.UUID `json:"reset_id"`
			Answers []string  `json:"answers"`
		}
		if err := httpx.Decode(r, &in, 4<<10); err != nil {
			return err
		}
		if err := s.ResetQuiz(r.Context(), in.ResetID, in.Answers); err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"sent": true})
		return nil
	}))
	r.Post("/auth/reset/complete", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		var in struct {
			Token       string `json:"token"`
			Code        string `json:"code"`
			NewPassword string `json:"new_password"`
		}
		if err := httpx.Decode(r, &in, 4<<10); err != nil {
			return err
		}
		if err := s.ResetComplete(r.Context(), in.Token, in.Code, in.NewPassword); err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"ok": true})
		return nil
	}))
}

// Private registers endpoints for signed-in callers (mount behind Middleware).
func (s *Service) Private(r chi.Router) {
	r.Post("/auth/logout", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.Must(r.Context())
		if err != nil {
			return err
		}
		if err := s.Logout(r.Context(), p.Session); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	// The app reports real use (a click or key press) here; background requests do not keep a session alive.
	r.Post("/auth/active", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.Must(r.Context())
		if err != nil {
			return err
		}
		if err := s.Active(r.Context(), p.Session, p.Org, p.Person); err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"idle_minutes": int(s.IdleLimit(r.Context(), p.Org) / time.Minute)})
		return nil
	}))
	// Where the inactivity clock stands, without counting as activity.
	r.Get("/auth/session", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.Must(r.Context())
		if err != nil {
			return err
		}
		lim := s.IdleLimit(r.Context(), p.Org)
		left := -1
		if lim > 0 {
			left = int((lim - time.Since(s.lastActive(r.Context(), p.Session))).Seconds())
		}
		httpx.JSON(w, 200, map[string]any{"idle_minutes": int(lim / time.Minute), "seconds_left": left})
		return nil
	}))
	r.Post("/auth/password", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.Must(r.Context())
		if err != nil {
			return err
		}
		var in struct {
			Current string `json:"current_password"`
			New     string `json:"new_password"`
		}
		if err := httpx.Decode(r, &in, 4<<10); err != nil {
			return err
		}
		if err := s.ChangePassword(r.Context(), p.Account, p.Session, in.Current, in.New, p.IP); err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"ok": true})
		return nil
	}))
	r.Post("/auth/mfa/setup", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.Must(r.Context())
		if err != nil {
			return err
		}
		sec, uri, err := s.MFASetup(r.Context(), p.Account)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"secret": sec, "otpauth_url": uri})
		return nil
	}))
	r.Post("/auth/mfa/confirm", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.Must(r.Context())
		if err != nil {
			return err
		}
		var in struct{ Code string }
		if err := httpx.Decode(r, &in, 1<<10); err != nil {
			return err
		}
		if err := s.MFAConfirm(r.Context(), p.Account, in.Code); err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"mfa_enabled": true})
		return nil
	}))
}

// DevOutbox exposes the dev mailer. Mount it only when HUB_ENV is not prod.
func DevOutbox(d *mail.Dev) http.HandlerFunc {
	return httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		httpx.JSON(w, 200, map[string]any{"messages": d.To(strings.ToLower(r.URL.Query().Get("to")))})
		return nil
	})
}

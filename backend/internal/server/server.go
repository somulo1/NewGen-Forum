// Package server assembles the HTTP API: middleware first, then every module's routes.
package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"hub/internal/app"
	"hub/internal/httpx"
	"hub/internal/mail"
	"hub/internal/modules/auth"
)

// Module mounts routes. Public ones need no sign-in; Private ones are mounted behind the auth middleware.
type Module interface {
	Public(chi.Router)
	Private(chi.Router)
}

type Deps struct {
	App     *app.App
	Auth    *auth.Service
	Dev     *mail.Dev
	Modules []Module
}

func Router(d Deps) http.Handler {
	r := chi.NewRouter()
	if err := httpx.SetTrustedProxies(d.App.Cfg.Proxies); err != nil {
		panic("HUB_TRUSTED_PROXIES: " + err.Error())
	}
	r.Use(httpx.RequestIDMW, httpx.Recover, httpx.SecurityHeaders, httpx.CORS(d.App.Cfg.AllowedOrigins), httpx.AccessLog, d.ipLimit, compressExceptSockets, httpx.BodyLimit(16<<20))
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) { httpx.JSON(w, 200, map[string]any{"ok": true}) })
	r.Get("/readyz", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		if err := d.App.DB.Pool.Ping(r.Context()); err != nil {
			return httpx.Err(503, "db_down", "database unavailable")
		}
		if err := d.App.Rdb.Ping(r.Context()).Err(); err != nil {
			return httpx.Err(503, "bus_down", "realtime bus unavailable")
		}
		httpx.JSON(w, 200, map[string]any{"ready": true, "connections": d.App.Hub.Connections()})
		return nil
	}))

	r.Route("/v1", func(v chi.Router) {
		d.Auth.Public(v)
		for _, m := range d.Modules {
			m.Public(v)
		}
		if d.Dev != nil && d.App.Cfg.Env != "prod" {
			v.Get("/dev/outbox", auth.DevOutbox(d.Dev))
		}
		v.Group(func(p chi.Router) {
			p.Use(d.Auth.Middleware, d.rateLimit)
			d.Auth.Private(p)
			for _, m := range d.Modules {
				m.Private(p)
			}
		})
	})
	if d.App.Cfg.DocsOn() {
		d.mountDocs(r)
	}
	return r
}

// compressExceptSockets gzips responses but leaves WebSocket upgrades alone (a compressing writer cannot be hijacked).
func compressExceptSockets(next http.Handler) http.Handler {
	gz := middleware.Compress(5)(next)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/realtime" {
			next.ServeHTTP(w, r)
			return
		}
		gz.ServeHTTP(w, r)
	})
}

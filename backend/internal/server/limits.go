package server

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"hub/internal/httpx"
	"hub/internal/tenant"
)

// Tier is a tighter cap on one expensive route, counted per signed-in account.
type Tier struct {
	Limit  int
	Window time.Duration
}

// Tiers are the routes that cost more than an ordinary request (they start calls, send mail, move files, fan out events).
// Every route is also under the per-account and per-address caps below. The API reference (/v1/docs) is built from this
// table, so what it says is what is enforced.
var Tiers = map[string]Tier{
	"POST /v1/calls":                      {30, time.Minute},
	"POST /v1/calls/{id}/join":            {60, time.Minute},
	"POST /v1/calls/{id}/control":         {120, time.Minute},
	"POST /v1/realtime/ticket":            {60, time.Minute},
	"POST /v1/conversations":              {30, time.Minute},
	"POST /v1/people":                     {120, time.Minute},
	"POST /v1/people/{id}/reset-password": {20, time.Minute},
	"POST /v1/tickets":                    {60, time.Minute},
	"POST /v1/comms":                      {30, time.Minute},
	"POST /v1/projects":                   {30, time.Minute},
	"POST /v1/files/{id}/shares":          {30, time.Minute},
	"POST /v1/folders/{id}/files":         {60, time.Minute},
	"POST /v1/e2ee/devices":               {10, time.Minute},
}

// publicTier: sign-in, sign-up, reset and invite routes need no token, so they are capped by address, hard.
func publicPath(p string) bool {
	return strings.HasPrefix(p, "/v1/auth/") || strings.HasPrefix(p, "/v1/invites/")
}

func limited(w http.ResponseWriter, retry time.Duration, limit int, msg string) {
	if retry < time.Second {
		retry = time.Second
	}
	w.Header().Set("Retry-After", strconv.Itoa(int(retry.Seconds())))
	w.Header().Set("RateLimit-Limit", strconv.Itoa(limit))
	w.Header().Set("RateLimit-Remaining", "0")
	httpx.JSON(w, 429, map[string]any{"error": httpx.Err(429, "rate_limited", msg)})
}

// ipLimit caps every address before anything else runs, and the public sign-in routes much harder. It runs ahead of
// authentication so a flood of bad requests costs one Redis counter, not a database lookup. Monitoring paths and the media
// servers' signed notices are not counted.
func (d Deps) ipLimit(next http.Handler) http.Handler {
	cfg := d.App.Cfg
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if p == "/healthz" || p == "/readyz" || p == "/v1/calls/sfu-hook" {
			next.ServeHTTP(w, r)
			return
		}
		ip := httpx.ClientIP(r)
		if publicPath(p) && r.Method != http.MethodOptions {
			if ok, retry, _ := d.App.Limit.Sliding(r.Context(), "ip-pub", ip, cfg.RateAuthIP, time.Minute, true); !ok {
				limited(w, retry, cfg.RateAuthIP, "Too many attempts. Wait a moment and try again.")
				return
			}
		}
		ok, retry, left := d.App.Limit.Sliding(r.Context(), "ip", ip, cfg.RateIP, time.Minute, false)
		if !ok {
			limited(w, retry, cfg.RateIP, "Slow down a little.")
			return
		}
		w.Header().Set("RateLimit-Limit", strconv.Itoa(cfg.RateIP))
		w.Header().Set("RateLimit-Remaining", strconv.Itoa(left))
		next.ServeHTTP(w, r)
	})
}

// rateLimit caps each signed-in account, and the expensive routes in Tiers on top (one bad client must not slow the others down).
func (d Deps) rateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p, ok := tenant.From(r.Context()); ok {
			rps := d.App.Cfg.RateRPS * 10
			if ok, retry, _ := d.App.Limit.Sliding(r.Context(), "acct", p.Account.String(), rps, 10*time.Second, false); !ok {
				limited(w, retry, rps, "Slow down a little.")
				return
			}
			if rc := chi.RouteContext(r.Context()); rc != nil {
				if t, ok := Tiers[r.Method+" "+rc.RoutePattern()]; ok {
					if ok, retry, _ := d.App.Limit.Sliding(r.Context(), "tier:"+r.Method+rc.RoutePattern(), p.Account.String(), t.Limit, t.Window, false); !ok {
						limited(w, retry, t.Limit, "You are doing that too often. Wait a moment and try again.")
						return
					}
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

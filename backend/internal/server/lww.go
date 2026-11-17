package server

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"hub/internal/app"
	"hub/internal/httpx"
	"hub/internal/tenant"
)

// Anything done offline in the phone app reaches the server later, carrying the time it was done (X-Hub-Written-At, Unix ms).
// Two rules hold for every endpoint:
//   - a change that waited more than 48 hours without a connection is refused (410 offline_expired): it is too old to apply;
//   - for a change to one thing (PUT, PATCH or DELETE on its address) the latest change by the time it was made wins: the server
//     remembers when the last applied change to each address was made, and an older one arriving late is not applied (409
//     stale_edit), so it never overwrites something changed after it. Changes made online count as made when they arrive.

const (
	offlineLimit = 48 * time.Hour
	lwwKeep      = offlineLimit + 24*time.Hour // longer than any change may wait
	lwwClockSkew = 2 * time.Minute
)

var (
	errStale   = httpx.Err(409, "stale_edit", "This was changed elsewhere after you made your change, so the newer change was kept.")
	errExpired = httpx.Err(410, "offline_expired", "This waited more than 48 hours without a connection, so it was not applied.")
)

// writtenAt is when the change was made (now when no time came with it, or one ahead of the clock), and whether it is too old.
func writtenAt(r *http.Request) (time.Time, bool) {
	now := time.Now()
	ms, err := strconv.ParseInt(strings.TrimSpace(r.Header.Get("X-Hub-Written-At")), 10, 64)
	if err != nil || ms <= 0 {
		return now, false
	}
	t := time.UnixMilli(ms)
	if t.After(now.Add(-lwwClockSkew)) {
		return now, false
	}
	return t, now.Sub(t) > offlineLimit
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) { w.status = code; w.ResponseWriter.WriteHeader(code) }

func latestWins(a *app.App) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}
			at, expired := writtenAt(r)
			if expired {
				httpx.Wrap(func(http.ResponseWriter, *http.Request) error { return errExpired })(w, r)
				return
			}
			p, ok := tenant.From(r.Context())
			if !ok || r.Method == http.MethodPost {
				next.ServeHTTP(w, r)
				return
			}
			key := "lww:" + p.Org.String() + ":" + r.URL.Path
			if last, err := a.Rdb.Get(r.Context(), key).Int64(); err == nil && at.UnixMilli() < last {
				httpx.Wrap(func(http.ResponseWriter, *http.Request) error { return errStale })(w, r)
				return
			}
			sw := &statusWriter{ResponseWriter: w, status: 200}
			next.ServeHTTP(sw, r)
			if sw.status < 300 {
				a.Rdb.Set(r.Context(), key, at.UnixMilli(), lwwKeep)
			}
		})
	}
}

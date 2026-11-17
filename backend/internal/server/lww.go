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

// Edits made offline in the phone app reach the server later, carrying the time they were made (X-Hub-Written-At, Unix ms).
// For a change to one thing (PUT, PATCH or DELETE on its address) the latest change by the time it was made wins: the server
// remembers when the last applied change to each address was made, and an older one arriving late is not applied (409
// stale_edit), so it never overwrites something changed after it. Changes made online count as made when they arrive.

const (
	lwwKeep      = 8 * 24 * time.Hour // longer than the phone keeps an edit waiting (7 days)
	lwwMaxAge    = 7 * 24 * time.Hour
	lwwClockSkew = 2 * time.Minute
)

var errStale = httpx.Err(409, "stale_edit", "This was changed elsewhere after you made your change, so the newer change was kept.")

// writtenAt is when the change was made: the header when it is believable, otherwise now.
func writtenAt(r *http.Request) time.Time {
	now := time.Now()
	ms, err := strconv.ParseInt(strings.TrimSpace(r.Header.Get("X-Hub-Written-At")), 10, 64)
	if err != nil || ms <= 0 {
		return now
	}
	t := time.UnixMilli(ms)
	if t.After(now.Add(lwwClockSkew)) || now.Sub(t) > lwwMaxAge {
		return now
	}
	if t.After(now) {
		return now
	}
	return t
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) { w.status = code; w.ResponseWriter.WriteHeader(code) }

func latestWins(a *app.App) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPut && r.Method != http.MethodPatch && r.Method != http.MethodDelete {
				next.ServeHTTP(w, r)
				return
			}
			p, ok := tenant.From(r.Context())
			if !ok {
				next.ServeHTTP(w, r)
				return
			}
			at := writtenAt(r)
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

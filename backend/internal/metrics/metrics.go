// Package metrics is a small Prometheus-format exporter: request counts and latency by route, gauges for the things that
// tell you a node is in trouble (open sockets, database pool, rooms), and a counter for panics and 5xx. It uses only the
// standard library so the API gains no dependency. Scrape it at /metrics with the token in HUB_METRICS_TOKEN.
package metrics

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5"
)

var buckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}

type series struct {
	count   atomic.Int64
	buckets [11]atomic.Int64
	sumUS   atomic.Int64
}

type key struct{ method, route, class string }

type Registry struct {
	mu     sync.RWMutex
	routes map[key]*series
	gauges map[string]func() float64
	inFlt  atomic.Int64
	start  time.Time
}

func New() *Registry {
	return &Registry{routes: map[key]*series{}, gauges: map[string]func() float64{}, start: time.Now()}
}

// Gauge registers a value read at scrape time, e.g. open sockets.
func (r *Registry) Gauge(name string, f func() float64) {
	r.mu.Lock()
	r.gauges[name] = f
	r.mu.Unlock()
}

type sw struct {
	http.ResponseWriter
	status int
}

func (s *sw) WriteHeader(c int)           { s.status = c; s.ResponseWriter.WriteHeader(c) }
func (s *sw) Unwrap() http.ResponseWriter { return s.ResponseWriter }

// Middleware counts every request by method, route pattern (never the raw path: ids would explode the series) and status class.
func (r *Registry) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		t := time.Now()
		r.inFlt.Add(1)
		s := &sw{ResponseWriter: w, status: 200}
		defer func() {
			r.inFlt.Add(-1)
			route := "unmatched"
			if rc := chi.RouteContext(req.Context()); rc != nil && rc.RoutePattern() != "" {
				route = rc.RoutePattern()
			}
			r.observe(key{req.Method, route, fmt.Sprintf("%dxx", s.status/100)}, time.Since(t))
		}()
		next.ServeHTTP(s, req)
	})
}

func (r *Registry) observe(k key, d time.Duration) {
	r.mu.RLock()
	s := r.routes[k]
	r.mu.RUnlock()
	if s == nil {
		r.mu.Lock()
		if s = r.routes[k]; s == nil {
			s = &series{}
			r.routes[k] = s
		}
		r.mu.Unlock()
	}
	sec := d.Seconds()
	s.count.Add(1)
	s.sumUS.Add(d.Microseconds())
	for i, b := range buckets {
		if sec <= b {
			s.buckets[i].Add(1)
		}
	}
}

// Handler serves the text format. A non-empty token is required as a bearer token; with none set only loopback may read it.
func (r *Registry) Handler(token string, clientIP func(*http.Request) string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if token != "" {
			if strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ") != token {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
		} else if ip := clientIP(req); ip != "127.0.0.1" && ip != "::1" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		var b strings.Builder
		fmt.Fprintf(&b, "# TYPE hub_uptime_seconds gauge\nhub_uptime_seconds %f\n# TYPE hub_requests_in_flight gauge\nhub_requests_in_flight %d\n", time.Since(r.start).Seconds(), r.inFlt.Load())
		r.mu.RLock()
		names := make([]string, 0, len(r.gauges))
		for n := range r.gauges {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, n := range names {
			fmt.Fprintf(&b, "# TYPE %s gauge\n%s %f\n", n, n, r.gauges[n]())
		}
		keys := make([]key, 0, len(r.routes))
		for k := range r.routes {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool { return fmt.Sprint(keys[i]) < fmt.Sprint(keys[j]) })
		b.WriteString("# TYPE hub_http_requests_total counter\n# TYPE hub_http_request_seconds histogram\n")
		for _, k := range keys {
			s := r.routes[k]
			l := fmt.Sprintf(`method="%s",route="%s",class="%s"`, k.method, k.route, k.class)
			fmt.Fprintf(&b, "hub_http_requests_total{%s} %d\n", l, s.count.Load())
			for i, bk := range buckets {
				fmt.Fprintf(&b, "hub_http_request_seconds_bucket{%s,le=\"%g\"} %d\n", l, bk, s.buckets[i].Load())
			}
			fmt.Fprintf(&b, "hub_http_request_seconds_bucket{%s,le=\"+Inf\"} %d\nhub_http_request_seconds_sum{%s} %f\nhub_http_request_seconds_count{%s} %d\n", l, s.count.Load(), l, float64(s.sumUS.Load())/1e6, l, s.count.Load())
		}
		r.mu.RUnlock()
		_, _ = w.Write([]byte(b.String()))
	})
}

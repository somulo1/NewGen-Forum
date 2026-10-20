// Package httpx holds small HTTP helpers shared by every module: JSON in/out, errors, middleware.
package httpx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ctxKey int

const reqIDKey ctxKey = 1

// Error is what handlers return; Wrap turns it into the JSON error body.
type Error struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string { return e.Code + ": " + e.Message }

func Err(status int, code, msg string) *Error { return &Error{status, code, msg} }

var (
	ErrUnauthorized = Err(401, "unauthorized", "Sign in to continue.")
	ErrForbidden    = Err(403, "forbidden", "You do not have the right to do that.")
	ErrNotFound     = Err(404, "not_found", "Not found.")
	ErrBadRequest   = Err(400, "bad_request", "That request is not valid.")
	ErrConflict     = Err(409, "conflict", "That conflicts with the current state.")
)

func Bad(msg string) *Error { return Err(400, "bad_request", msg) }

// Handler is a handler that returns an error instead of writing it.
type Handler func(w http.ResponseWriter, r *http.Request) error

// Wrap runs h and writes any returned error as JSON. Unexpected errors are logged but never shown to the caller.
func Wrap(h Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			var he *Error
			if errors.As(err, &he) {
				JSON(w, he.Status, map[string]any{"error": he})
				return
			}
			slog.Error("request failed", "err", err, "path", r.URL.Path, "req", RequestID(r.Context()))
			JSON(w, 500, map[string]any{"error": &Error{500, "internal", "Something went wrong. Try again."}})
		}
	}
}

func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

// Decode reads a JSON body with a size cap and refuses unknown fields and trailing data.
func Decode(r *http.Request, dst any, maxBytes int64) error {
	if maxBytes <= 0 {
		maxBytes = 1 << 20
	}
	dec := json.NewDecoder(io.LimitReader(r.Body, maxBytes+1))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return Bad("The request body is not valid JSON for this endpoint.")
	}
	if dec.More() {
		return Bad("Unexpected data after the request body.")
	}
	return nil
}

func RequestID(ctx context.Context) string { s, _ := ctx.Value(reqIDKey).(string); return s }

// ClientIP prefers the first hop set by a trusted proxy; direct callers fall back to the socket address.
var trustedProxies []*net.IPNet

// SetTrustedProxies names the proxies (CIDRs) whose X-Forwarded-For header may be believed. Anyone else can write that header
// themselves, so without this an attacker would pick their own "address" and walk around every per-address limit.
func SetTrustedProxies(cidrs []string) error {
	var out []*net.IPNet
	for _, c := range cidrs {
		if c = strings.TrimSpace(c); c == "" {
			continue
		}
		if !strings.Contains(c, "/") {
			if strings.Contains(c, ":") {
				c += "/128"
			} else {
				c += "/32"
			}
		}
		_, n, err := net.ParseCIDR(c)
		if err != nil {
			return err
		}
		out = append(out, n)
	}
	trustedProxies = out
	return nil
}

func trusted(ip string) bool {
	a := net.ParseIP(ip)
	if a == nil {
		return false
	}
	for _, n := range trustedProxies {
		if n.Contains(a) {
			return true
		}
	}
	return false
}

// ClientIP is the caller's address: the connection's, unless it comes from a trusted proxy, in which case it is the nearest
// address in X-Forwarded-For that is not itself a trusted proxy.
func ClientIP(r *http.Request) string {
	h, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		h = r.RemoteAddr
	}
	if !trusted(h) {
		return h
	}
	f := r.Header.Get("X-Forwarded-For")
	if f == "" {
		return h
	}
	parts := strings.Split(f, ",")
	for i := len(parts) - 1; i >= 0; i-- {
		if ip := strings.TrimSpace(parts[i]); ip != "" && !trusted(ip) {
			return ip
		}
	}
	return strings.TrimSpace(parts[0])
}

// --- middleware ---

func RequestIDMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.NewString()
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), reqIDKey, id)))
	})
}

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				slog.Error("panic", "value", fmt.Sprint(v), "stack", string(debug.Stack()), "req", RequestID(r.Context()))
				JSON(w, 500, map[string]any{"error": &Error{500, "internal", "Something went wrong. Try again."}})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (s *statusWriter) WriteHeader(c int)           { s.status = c; s.ResponseWriter.WriteHeader(c) }
func (s *statusWriter) Unwrap() http.ResponseWriter { return s.ResponseWriter }

func AccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(sw, r)
		slog.Info("http", "m", r.Method, "p", r.URL.Path, "s", sw.status, "ms", time.Since(t).Milliseconds(), "req", RequestID(r.Context()))
	})
}

// SecurityHeaders applies strict defaults for an API.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Cache-Control", "no-store")
		h.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}

// CORS allows only the listed origins and never combines wildcard with credentials.
func CORS(origins []string) func(http.Handler) http.Handler {
	ok := map[string]bool{}
	for _, o := range origins {
		ok[strings.TrimSpace(o)] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if o := r.Header.Get("Origin"); o != "" && ok[o] {
				h := w.Header()
				h.Set("Access-Control-Allow-Origin", o)
				h.Set("Vary", "Origin")
				h.Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Org")
				h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				h.Set("Access-Control-Max-Age", "600")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// BodyLimit caps every request body.
func BodyLimit(n int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// file uploads carry their own, larger limit (set by the upload handler from configuration)
			if !(r.Method == http.MethodPut && strings.HasSuffix(r.URL.Path, "/content")) {
				r.Body = http.MaxBytesReader(w, r.Body, n)
			}
			next.ServeHTTP(w, r)
		})
	}
}

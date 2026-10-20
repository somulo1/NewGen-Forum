package server_test

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"hub/internal/httpx"
	"hub/internal/testkit"
)

func get(t *testing.T, url string, hdr map[string]string) (int, http.Header, []byte) {
	t.Helper()
	req, _ := http.NewRequest("GET", url, nil)
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	return res.StatusCode, res.Header, b
}

// The reference is built from the router, so a route that exists must be in it, with its auth and rate limit.
func TestAPIReferenceListsEveryRoute(t *testing.T) {
	e := testkit.New(t)
	code, _, b := get(t, e.Srv.URL+"/v1/docs/openapi.json", nil)
	if code != 200 {
		t.Fatalf("openapi.json: %d", code)
	}
	var spec struct {
		Paths map[string]map[string]struct {
			Security  []any          `json:"security"`
			Limit     string         `json:"x-rate-limit"`
			Responses map[string]any `json:"responses"`
		} `json:"paths"`
	}
	if err := json.Unmarshal(b, &spec); err != nil {
		t.Fatal(err)
	}
	for _, r := range []struct{ method, path string }{
		{"post", "/v1/auth/login"}, {"get", "/v1/conversations"}, {"post", "/v1/conversations/{id}/messages"}, {"post", "/v1/calls/{id}/control"},
		{"get", "/v1/tickets"}, {"post", "/v1/people"}, {"get", "/healthz"},
	} {
		op, ok := spec.Paths[r.path][r.method]
		if !ok {
			t.Fatalf("%s %s is missing from the reference", r.method, r.path)
		}
		if op.Limit == "" || op.Responses["429"] == nil || op.Responses["401"] == nil {
			t.Fatalf("%s %s lacks rate limit or error docs: %+v", r.method, r.path, op)
		}
	}
	if op := spec.Paths["/v1/auth/login"]["post"]; op.Security == nil || len(op.Security) != 0 {
		t.Fatalf("sign-in must be documented as needing no token: %+v", op.Security)
	}
	if op := spec.Paths["/v1/conversations"]["get"]; len(op.Security) != 0 {
		t.Fatalf("conversations must need a token: %+v", op.Security)
	}
	if !strings.Contains(spec.Paths["/v1/calls"]["post"].Limit, "30 per 1m0s") {
		t.Fatalf("the documented limit must be the enforced one: %s", spec.Paths["/v1/calls"]["post"].Limit)
	}
	if n := len(spec.Paths); n < 100 {
		t.Fatalf("only %d paths documented", n)
	}
	if code, _, page := get(t, e.Srv.URL+"/v1/docs", nil); code != 200 || !strings.Contains(string(page), "Hub API reference") {
		t.Fatalf("docs page: %d", code)
	}
}

// Public routes are capped by address, and an address cannot be chosen by the caller.
func TestAddressLimitsAndSpoofedForwardingHeaders(t *testing.T) {
	e := testkit.New(t)
	url := e.Srv.URL + "/v1/auth/refresh"
	post := func(xff string) (int, http.Header) {
		req, _ := http.NewRequest("POST", url, strings.NewReader(`{"refresh_token":"nope"}`))
		req.Header.Set("Content-Type", "application/json")
		if xff != "" {
			req.Header.Set("X-Forwarded-For", xff)
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		return res.StatusCode, res.Header
	}
	limit := e.A.Cfg.RateAuthIP
	hit := 0
	var last int
	var hdr http.Header
	for i := 0; i < limit+5; i++ {
		last, hdr = post("203.0.113.77")
		if last == 429 {
			break
		}
		hit++
	}
	if last != 429 || hit > limit || hdr.Get("Retry-After") == "" {
		t.Fatalf("expected a 429 with Retry-After after %d hits, got %d after %d (%v)", limit, last, hit, hdr)
	}
	// a different address is not affected
	if code, _ := post("203.0.113.78"); code == 429 {
		t.Fatalf("one address blocked another")
	}
}

func TestForwardedForOnlyBelievedFromTrustedProxies(t *testing.T) {
	if err := httpx.SetTrustedProxies([]string{"10.0.0.0/8"}); err != nil {
		t.Fatal(err)
	}
	defer httpx.SetTrustedProxies([]string{"127.0.0.0/8", "::1/128"})
	mk := func(remote, xff string) *http.Request {
		r, _ := http.NewRequest("GET", "/", nil)
		r.RemoteAddr = remote
		if xff != "" {
			r.Header.Set("X-Forwarded-For", xff)
		}
		return r
	}
	if ip := httpx.ClientIP(mk("198.51.100.9:5555", "1.2.3.4")); ip != "198.51.100.9" {
		t.Fatalf("a stranger chose their own address: %s", ip)
	}
	if ip := httpx.ClientIP(mk("10.0.0.5:5555", "1.2.3.4")); ip != "1.2.3.4" {
		t.Fatalf("the proxy's forwarded address was ignored: %s", ip)
	}
	// the client prepends a fake address; the proxy appends the real one: the real one wins
	if ip := httpx.ClientIP(mk("10.0.0.5:5555", "9.9.9.9, 1.2.3.4")); ip != "1.2.3.4" {
		t.Fatalf("a forged entry was believed: %s", ip)
	}
}

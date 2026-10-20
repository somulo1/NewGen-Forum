package server_test

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"hub/internal/config"
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

// The reference is built from the router and every route needs a written entry, so a new route cannot ship undocumented.
func TestAPIReferenceDescribesEveryRoute(t *testing.T) {
	e := testkit.New(t)
	code, _, b := get(t, e.Srv.URL+"/v1/docs/openapi.json", nil)
	if code != 200 {
		t.Fatalf("openapi.json: %d", code)
	}
	var spec struct {
		Paths map[string]map[string]struct {
			Summary    string         `json:"summary"`
			Documented bool           `json:"x-documented"`
			Security   []any          `json:"security"`
			Responses  map[string]any `json:"responses"`
			Body       any            `json:"requestBody"`
		} `json:"paths"`
	}
	if err := json.Unmarshal(b, &spec); err != nil {
		t.Fatal(err)
	}
	ops := 0
	for path, ms := range spec.Paths {
		for method, op := range ms {
			ops++
			if !op.Documented {
				t.Errorf("%s %s has no entry in internal/server/apidocs", strings.ToUpper(method), path)
			}
			if strings.Contains(strings.ToLower(op.Summary+path), "rate limit") {
				t.Errorf("the reference should not describe rate limits: %s", path)
			}
		}
	}
	if ops < 100 {
		t.Fatalf("only %d operations documented", ops)
	}
	if op := spec.Paths["/v1/auth/login"]["post"]; op.Security == nil || len(op.Security) != 0 || op.Body == nil {
		t.Fatalf("sign-in must be documented as needing no token and taking a body: %+v", op)
	}
	if op := spec.Paths["/v1/conversations"]["get"]; len(op.Security) != 0 {
		t.Fatalf("conversations must need a token: %+v", op.Security)
	}
	if code, _, page := get(t, e.Srv.URL+"/v1/docs", nil); code != 200 || !strings.Contains(string(page), "Hub API reference") {
		t.Fatalf("docs page: %d", code)
	}
}

func TestDocsOnlyInConfiguredEnvironments(t *testing.T) {
	for env, listed := range map[string]string{"dev": "dev,staging,prod", "staging": "dev,staging,prod", "prod": "dev,staging,prod", "production": "prod", "development": "dev", "prod ": "dev,staging"} {
		c := &config.Config{Env: env, DocsEnvs: strings.Split(listed, ",")}
		want := strings.Contains(","+listed+",", ","+strings.TrimSpace(env)+",") || env == "production" && strings.Contains(listed, "prod") || env == "development" && strings.Contains(listed, "dev")
		if c.DocsOn() != want {
			t.Errorf("env %q with HUB_DOCS_ENVS=%q: DocsOn=%v want %v", env, listed, c.DocsOn(), want)
		}
	}
	if (&config.Config{Env: "prod", DocsEnvs: []string{"dev", "staging"}}).DocsOn() {
		t.Fatalf("docs must be off in an environment that is not listed")
	}
	if (&config.Config{Env: "staging", DocsEnvs: nil}).DocsOn() {
		t.Fatalf("an empty list turns docs off everywhere")
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

// Row-level security is declared on the parent tables, and PostgreSQL ignores it for a query that names a partition. The
// application role must therefore have no direct access to any partition, or one injected query could read every organization.
func TestApplicationRoleCannotReadPartitionsDirectly(t *testing.T) {
	e := testkit.New(t)
	ctx := t.Context()
	rows, err := e.A.DB.Pool.Query(ctx, `select c.relname from pg_class c join pg_namespace n on n.oid = c.relnamespace where n.nspname = 'hub' and c.relispartition and c.relkind = 'r'`)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for rows.Next() {
		var n string
		_ = rows.Scan(&n)
		names = append(names, n)
	}
	rows.Close()
	if len(names) < 40 {
		t.Fatalf("expected the messages and records partitions, found %d", len(names))
	}
	for _, n := range names {
		var c int
		if err := e.A.DB.Pool.QueryRow(ctx, `select count(*) from hub.`+n).Scan(&c); err == nil {
			t.Fatalf("the application role can read partition %s directly, bypassing row-level security", n)
		}
	}
}

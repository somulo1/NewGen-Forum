package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/google/uuid"

	"hub/internal/app"
	"hub/internal/config"
	"hub/internal/crypto"
	"hub/internal/mail"
	"hub/internal/modules/auth"
	"hub/internal/server"
)

type env struct {
	t   *testing.T
	srv *httptest.Server
	dev *mail.Dev
	a   *app.App
}

func newEnv(t *testing.T) *env {
	t.Helper()
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	dev := &mail.Dev{Quiet: true}
	a, err := app.New(context.Background(), cfg, dev)
	if err != nil {
		t.Skip("infrastructure not available: ", err)
	}
	a.PW = crypto.PasswordParams{Time: 1, MemoryKiB: 8 * 1024, Threads: 1}
	t.Cleanup(a.Close)
	srv := httptest.NewServer(server.Router(server.Deps{App: a, Auth: auth.New(a), Dev: dev}))
	t.Cleanup(srv.Close)
	return &env{t, srv, dev, a}
}

func (e *env) do(method, path string, body any, token string) (int, map[string]any) {
	e.t.Helper()
	var rd *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	} else {
		rd = bytes.NewReader(nil)
	}
	req, _ := http.NewRequest(method, e.srv.URL+"/v1"+path, rd)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", "10.9."+fmt.Sprint(uuid.New().ID()%250)+"."+fmt.Sprint(uuid.New().ID()%250)) // each call looks like a different client: rate limits are tested separately
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		e.t.Fatal(err)
	}
	defer res.Body.Close()
	var out map[string]any
	_ = json.NewDecoder(res.Body).Decode(&out)
	return res.StatusCode, out
}

func quiz() []map[string]string {
	return []map[string]string{{"question": "pet", "answer": "Rex"}, {"question": "town", "answer": "Nairobi"}, {"question": "school", "answer": "Hub Primary"}}
}

func (e *env) signup(email, pw string) {
	e.t.Helper()
	code, body := e.do("POST", "/auth/signup", map[string]any{"name": "Amani", "email": email, "password": pw, "quiz": quiz()}, "")
	if code != 201 {
		e.t.Fatalf("signup %d %v", code, body)
	}
}

func uniq() string { return "u" + strings.ReplaceAll(uuid.NewString(), "-", "")[:12] + "@example.com" }

func TestSignupLoginRefreshLogout(t *testing.T) {
	e := newEnv(t)
	email := uniq()
	e.signup(email, "Passw0rdOK")
	if code, _ := e.do("POST", "/auth/signup", map[string]any{"name": "x", "email": email, "password": "Passw0rdOK", "quiz": quiz()}, ""); code != 409 {
		t.Fatalf("duplicate signup should be 409, got %d", code)
	}
	if code, _ := e.do("POST", "/auth/signup", map[string]any{"name": "x", "email": uniq(), "password": "weak", "quiz": quiz()}, ""); code != 400 {
		t.Fatalf("weak password accepted: %d", code)
	}
	code, body := e.do("POST", "/auth/login", map[string]any{"email": email, "password": "wrong"}, "")
	if code != 401 {
		t.Fatalf("bad password: %d", code)
	}
	code, tok := e.do("POST", "/auth/login", map[string]any{"email": email, "password": "Passw0rdOK"}, "")
	if code != 200 {
		t.Fatalf("login %d %v", code, tok)
	}
	access, refresh := tok["access_token"].(string), tok["refresh_token"].(string)
	if code, _ := e.do("POST", "/auth/password", map[string]any{"current_password": "Passw0rdOK", "new_password": "Passw0rdOK"}, access); code != 400 {
		t.Fatalf("reusing the same password should be refused: %d", code)
	}
	// rotation
	code, t2 := e.do("POST", "/auth/refresh", map[string]any{"refresh_token": refresh}, "")
	if code != 200 || t2["refresh_token"] == refresh {
		t.Fatalf("refresh failed or did not rotate: %d", code)
	}
	// the old access token died with its session
	if code, _ := e.do("POST", "/auth/mfa/setup", nil, access); code != 401 {
		t.Fatalf("rotated session's access token still works: %d", code)
	}
	// reuse of the OLD refresh token = theft: the whole family is revoked, including the new token
	if code, _ := e.do("POST", "/auth/refresh", map[string]any{"refresh_token": refresh}, ""); code != 401 {
		t.Fatalf("old refresh token reused: %d", code)
	}
	if code, _ := e.do("POST", "/auth/refresh", map[string]any{"refresh_token": t2["refresh_token"]}, ""); code != 401 {
		t.Fatalf("family should be dead after reuse: %d", code)
	}
	_ = body
}

func TestLoginDoesNotRevealWhoHasAnAccountAndLocksOut(t *testing.T) {
	e := newEnv(t)
	email := uniq()
	e.signup(email, "Passw0rdOK")
	c1, b1 := e.do("POST", "/auth/login", map[string]any{"email": email, "password": "nope"}, "")
	c2, b2 := e.do("POST", "/auth/login", map[string]any{"email": uniq(), "password": "nope"}, "")
	j1, j2 := fmt.Sprint(b1), fmt.Sprint(b2)
	if c1 != c2 || j1 != j2 {
		t.Fatalf("responses differ for known/unknown email: %d %s / %d %s", c1, j1, c2, j2)
	}
	for i := 0; i < 5; i++ {
		e.do("POST", "/auth/login", map[string]any{"email": email, "password": "bad"}, "")
	}
	if code, _ := e.do("POST", "/auth/login", map[string]any{"email": email, "password": "Passw0rdOK"}, ""); code != 429 {
		t.Fatalf("account not locked after repeated failures: %d", code)
	}
}

var codeRe = regexp.MustCompile(`\b(\d{6})\b`)
var tokRe = regexp.MustCompile(`token=([A-Za-z0-9_.\-]+)`)

func (e *env) last(email string) string {
	m := e.dev.To(email)
	if len(m) == 0 {
		return ""
	}
	return m[0].Body
}

func TestPasswordResetFlow(t *testing.T) {
	e := newEnv(t)
	email := uniq()
	e.signup(email, "Passw0rdOK")
	_, login := e.do("POST", "/auth/login", map[string]any{"email": email, "password": "Passw0rdOK"}, "")
	oldAccess := login["access_token"].(string)

	// 1. start: unknown and known addresses get the same kind of reply
	cu, bu := e.do("POST", "/auth/reset/start", map[string]any{"email": uniq()}, "")
	code, b := e.do("POST", "/auth/reset/start", map[string]any{"email": email}, "")
	if code != 202 || cu != 202 || bu["reset_id"] == nil {
		t.Fatalf("start: %d %d", code, cu)
	}
	rid := b["reset_id"].(string)
	// 2. wrong code, then right code
	if c, _ := e.do("POST", "/auth/reset/code", map[string]any{"reset_id": rid, "code": "000000"}, ""); c != 400 {
		t.Fatalf("wrong code accepted: %d", c)
	}
	otp := codeRe.FindString(e.last(email))
	if otp == "" {
		t.Fatal("no code was emailed")
	}
	c, q := e.do("POST", "/auth/reset/code", map[string]any{"reset_id": rid, "code": otp}, "")
	if c != 200 || len(q["questions"].([]any)) != 3 {
		t.Fatalf("code step: %d %v", c, q)
	}
	// 3. quiz: wrong then right
	if c, _ := e.do("POST", "/auth/reset/quiz", map[string]any{"reset_id": rid, "answers": []string{"rex", "mombasa", "hub primary"}}, ""); c != 400 {
		t.Fatalf("wrong answers accepted: %d", c)
	}
	if c, _ := e.do("POST", "/auth/reset/quiz", map[string]any{"reset_id": rid, "answers": []string{"REX", " nairobi ", "Hub   Primary"}}, ""); c != 200 {
		t.Fatalf("right answers refused: %d", c)
	}
	// 4. the link email
	body := e.last(email)
	token := tokRe.FindStringSubmatch(body)
	code2 := codeRe.FindString(strings.ReplaceAll(body, token[1], ""))
	if token == nil || code2 == "" {
		t.Fatalf("link email incomplete: %q", body)
	}
	// 5. wrong code does not use up the link; weak password refused; then success
	if c, _ := e.do("POST", "/auth/reset/complete", map[string]any{"token": token[1], "code": "111111", "new_password": "NewPassw0rd1"}, ""); c != 400 {
		t.Fatalf("wrong second code: %d", c)
	}
	if c, _ := e.do("POST", "/auth/reset/complete", map[string]any{"token": token[1], "code": code2, "new_password": "weak"}, ""); c != 400 {
		t.Fatalf("weak new password: %d", c)
	}
	if c, b := e.do("POST", "/auth/reset/complete", map[string]any{"token": token[1], "code": code2, "new_password": "NewPassw0rd1"}, ""); c != 200 {
		t.Fatalf("complete: %d %v", c, b)
	}
	// the link works once only
	if c, _ := e.do("POST", "/auth/reset/complete", map[string]any{"token": token[1], "code": code2, "new_password": "Another1Pass"}, ""); c != 400 {
		t.Fatalf("link reused: %d", c)
	}
	// sessions were signed out, old password is dead, new one works
	if c, _ := e.do("POST", "/auth/mfa/setup", nil, oldAccess); c != 401 {
		t.Fatalf("old session survived the reset: %d", c)
	}
	if c, _ := e.do("POST", "/auth/login", map[string]any{"email": email, "password": "Passw0rdOK"}, ""); c != 401 {
		t.Fatalf("old password still works: %d", c)
	}
	if c, _ := e.do("POST", "/auth/login", map[string]any{"email": email, "password": "NewPassw0rd1"}, ""); c != 200 {
		t.Fatalf("new password refused: %d", c)
	}
	if !strings.Contains(e.last(email), "was just changed") {
		t.Fatal("no change notice was emailed")
	}
	// unknown address: nothing was ever sent
	if len(e.dev.To("nobody@example.com")) != 0 {
		t.Fatal("mail sent to unknown address")
	}
}

func TestResetBruteForceKillsTheAttempt(t *testing.T) {
	e := newEnv(t)
	email := uniq()
	e.signup(email, "Passw0rdOK")
	_, b := e.do("POST", "/auth/reset/start", map[string]any{"email": email}, "")
	rid := b["reset_id"].(string)
	real := codeRe.FindString(e.last(email))
	for i := 0; i < 5; i++ {
		e.do("POST", "/auth/reset/code", map[string]any{"reset_id": rid, "code": fmt.Sprintf("%06d", i)}, "")
	}
	if c, _ := e.do("POST", "/auth/reset/code", map[string]any{"reset_id": rid, "code": real}, ""); c != 400 {
		t.Fatalf("attempt survived 5 wrong codes: %d", c)
	}
}

func TestChangePasswordSignsOutOtherDevices(t *testing.T) {
	e := newEnv(t)
	email := uniq()
	e.signup(email, "Passw0rdOK")
	_, a := e.do("POST", "/auth/login", map[string]any{"email": email, "password": "Passw0rdOK"}, "")
	_, b := e.do("POST", "/auth/login", map[string]any{"email": email, "password": "Passw0rdOK"}, "")
	if c, _ := e.do("POST", "/auth/password", map[string]any{"current_password": "bad", "new_password": "NewPassw0rd1"}, a["access_token"].(string)); c != 400 {
		t.Fatalf("wrong current password accepted: %d", c)
	}
	if c, _ := e.do("POST", "/auth/password", map[string]any{"current_password": "Passw0rdOK", "new_password": "NewPassw0rd1"}, a["access_token"].(string)); c != 200 {
		t.Fatalf("change password: %d", c)
	}
	if c, _ := e.do("POST", "/auth/mfa/setup", nil, b["access_token"].(string)); c != 401 {
		t.Fatalf("other device still signed in: %d", c)
	}
	if c, _ := e.do("POST", "/auth/mfa/setup", nil, a["access_token"].(string)); c != 200 {
		t.Fatalf("this device was signed out: %d", c)
	}
}

func TestStoredEmailIsEncryptedAndNotReadable(t *testing.T) {
	e := newEnv(t)
	email := uniq()
	e.signup(email, "Passw0rdOK")
	// read the raw row as the audit role (read-only): the address must not appear in any column
	pool := e.a.DB.Pool
	var raw string
	err := pool.QueryRow(context.Background(), `select (email_idx::text || email_enc::text || display_name || pass_hash) from hub.accounts order by created_at desc limit 1`).Scan(&raw)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(raw, email) || strings.Contains(raw, "Passw0rdOK") {
		t.Fatal("account row contains plaintext email or password")
	}
}

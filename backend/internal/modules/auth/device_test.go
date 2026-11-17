package auth_test

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

// phone stands in for the Hub phone app: it keeps a private key and signs what the server asks for.
type phone struct{ priv ed25519.PrivateKey }

func newPhone() phone { _, k, _ := ed25519.GenerateKey(rand.Reader); return phone{k} }

func (p phone) sign(msg string, at time.Time) (int64, string) {
	ts := at.Unix()
	return ts, base64.StdEncoding.EncodeToString(ed25519.Sign(p.priv, []byte(msg+"|"+strconv.FormatInt(ts, 10))))
}

func (p phone) login(email string) map[string]any {
	ts, sig := p.sign("hub-device-login|"+strings.ToLower(email), time.Now())
	return map[string]any{"key": base64.StdEncoding.EncodeToString(p.priv.Public().(ed25519.PublicKey)), "ts": ts, "sig": sig, "platform": "android 36"}
}

// refresh sends a refresh with (or, with no phone, without) the device signature headers.
func (e *env) refresh(token string, p *phone, at time.Time) (int, map[string]any) {
	e.t.Helper()
	b, _ := json.Marshal(map[string]any{"refresh_token": token})
	req, _ := http.NewRequest("POST", e.srv.URL+"/v1/auth/refresh", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if p != nil {
		ts, sig := p.sign("hub-device-refresh|"+strings.SplitN(token, ".", 2)[0], at)
		req.Header.Set("X-Hub-Device-TS", strconv.FormatInt(ts, 10))
		req.Header.Set("X-Hub-Device-Sig", sig)
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

func TestPhoneSignInIsBoundToThePhone(t *testing.T) {
	e := newEnv(t)
	email := uniq()
	e.signup(email, "Passw0rdOK")
	me, other := newPhone(), newPhone()

	// a proof signed by a different key than the one sent is refused before anything else
	bad := me.login(email)
	bad["key"] = other.login(email)["key"]
	if code, _ := e.do("POST", "/auth/login", map[string]any{"email": email, "password": "Passw0rdOK", "device": bad}, ""); code != 400 {
		t.Fatalf("forged phone proof accepted: %d", code)
	}

	tok := e.phoneLogin(email, me)
	r1 := tok["refresh_token"].(string)

	// the phone renews with its own signature
	code, t2 := e.refresh(r1, &me, time.Now())
	if code != 200 {
		t.Fatalf("signed refresh %d %v", code, t2)
	}
	r2 := t2["refresh_token"].(string)

	// another phone's key, or an old signature, does not renew it
	if code, _ := e.refresh(r2, &other, time.Now()); code != 401 {
		t.Fatalf("refresh signed by another phone: %d", code)
	}
	code, tok = e.do("POST", "/auth/login", map[string]any{"email": email, "password": "Passw0rdOK", "device": me.login(email)}, "")
	if code != 200 {
		t.Fatalf("login again %d", code)
	}
	r3 := tok["refresh_token"].(string)
	if code, _ := e.refresh(r3, &me, time.Now().Add(-10*time.Minute)); code != 401 {
		t.Fatalf("stale signature accepted: %d", code)
	}

	// a copied refresh token without the phone ends the whole sign-in: even the phone cannot use what follows from it
	code, tok = e.do("POST", "/auth/login", map[string]any{"email": email, "password": "Passw0rdOK", "device": me.login(email)}, "")
	if code != 200 {
		t.Fatalf("login again %d", code)
	}
	r4 := tok["refresh_token"].(string)
	code, body := e.refresh(r4, nil, time.Now())
	if code != 401 || !strings.Contains(strings.ToLower(jsonText(body)), "device_unverified") {
		t.Fatalf("unsigned refresh of a bound sign-in: %d %v", code, body)
	}
	if code, _ := e.refresh(r4, &me, time.Now()); code != 401 {
		t.Fatalf("the sign-in should have ended: %d", code)
	}

	// a browser sign-in (no phone) keeps working without signatures
	code, tok = e.do("POST", "/auth/login", map[string]any{"email": email, "password": "Passw0rdOK"}, "")
	if code != 200 {
		t.Fatalf("browser login %d", code)
	}
	if code, _ := e.refresh(tok["refresh_token"].(string), nil, time.Now()); code != 200 {
		t.Fatalf("browser refresh %d", code)
	}
}

func jsonText(v any) string { b, _ := json.Marshal(v); return string(b) }

func errCode(out map[string]any) string {
	if e, ok := out["error"].(map[string]any); ok {
		c, _ := e["code"].(string)
		return c
	}
	return ""
}

func (e *env) phoneMails(email string) []string {
	var out []string
	for _, m := range e.dev.To(email) {
		if strings.Contains(m.Subject, "Confirm your phone") {
			out = append(out, m.Body)
		}
	}
	return out
}

// phoneLogin signs in from a phone the account has not confirmed: the first try emails a code, the second (with it) works.
func (e *env) phoneLogin(email string, p phone) map[string]any {
	e.t.Helper()
	before := len(e.phoneMails(email))
	code, out := e.do("POST", "/auth/login", map[string]any{"email": email, "password": "Passw0rdOK", "device": p.login(email)}, "")
	if code != 401 || errCode(out) != "device_confirm" {
		e.t.Fatalf("a new phone must be confirmed first: %d %v", code, out)
	}
	var mails []string
	for i := 0; i < 100 && len(mails) <= before; i++ {
		time.Sleep(20 * time.Millisecond)
		mails = e.phoneMails(email)
	}
	if len(mails) <= before {
		e.t.Fatal("no confirmation email")
	}
	c := codeRe.FindStringSubmatch(mails[len(mails)-1])[1]
	code, tok := e.do("POST", "/auth/login", map[string]any{"email": email, "password": "Passw0rdOK", "device": p.login(email), "device_code": c}, "")
	if code != 200 {
		e.t.Fatalf("confirmed phone login %d %v", code, tok)
	}
	return tok
}

// One browser and one phone: each new sign-in replaces only the previous one of its kind, and a new phone needs the emailed
// code, after which the old phone is signed out and needs the same confirmation to come back.
func TestOneBrowserAndOnePhone(t *testing.T) {
	e := newEnv(t)
	email := uniq()
	e.signup(email, "Passw0rdOK")
	alive := func(access string) bool { c, _ := e.do("GET", "/auth/session", nil, access); return c == 200 }

	_, webA := e.do("POST", "/auth/login", map[string]any{"email": email, "password": "Passw0rdOK"}, "")
	first, second := newPhone(), newPhone()
	p1 := e.phoneLogin(email, first)
	if !alive(webA["access_token"].(string)) {
		t.Fatal("signing in on the phone signed the browser out")
	}

	_, webB := e.do("POST", "/auth/login", map[string]any{"email": email, "password": "Passw0rdOK"}, "")
	if c, out := e.do("GET", "/auth/session", nil, webA["access_token"].(string)); c != 401 || errCode(out) != "session_replaced" {
		t.Fatalf("the first browser must be signed out by the second: %d %v", c, out)
	}
	code, p1b := e.refresh(p1["refresh_token"].(string), &first, time.Now())
	if code != 200 {
		t.Fatalf("a browser sign-in signed the phone out: %d", code)
	}

	// another phone: nothing changes until it is confirmed; a wrong code is refused
	if code, out := e.do("POST", "/auth/login", map[string]any{"email": email, "password": "Passw0rdOK", "device": second.login(email), "device_code": "000000"}, ""); code != 401 || (errCode(out) != "device_code" && errCode(out) != "device_code_expired") {
		t.Fatalf("a made-up code: %d %v", code, out)
	}
	if !alive(p1b["access_token"].(string)) {
		t.Fatal("an unconfirmed phone signed the first phone out")
	}
	e.phoneLogin(email, second)
	code, out := e.refresh(p1b["refresh_token"].(string), &first, time.Now())
	if code != 401 || errCode(out) != "phone_replaced" {
		t.Fatalf("the old phone must be signed out once the new one is confirmed: %d %v", code, out)
	}
	if !alive(webB["access_token"].(string)) {
		t.Fatal("confirming a new phone signed the browser out")
	}
	// the old phone comes back only through the same confirmation
	if code, out := e.do("POST", "/auth/login", map[string]any{"email": email, "password": "Passw0rdOK", "device": first.login(email)}, ""); code != 401 || errCode(out) != "device_confirm" {
		t.Fatalf("the replaced phone signed in without confirmation: %d %v", code, out)
	}
}

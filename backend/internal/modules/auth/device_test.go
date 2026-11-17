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

	code, tok := e.do("POST", "/auth/login", map[string]any{"email": email, "password": "Passw0rdOK", "device": me.login(email)}, "")
	if code != 200 {
		t.Fatalf("phone login %d %v", code, tok)
	}
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

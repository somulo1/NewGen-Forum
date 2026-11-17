package push

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"

	"hub/internal/realtime"
)

func pemOf(t *testing.T, typ string, der []byte) string {
	t.Helper()
	return string(pem.EncodeToMemory(&pem.Block{Type: typ, Bytes: der}))
}

func TestFCMSignsInSendsAndDropsDeadAddresses(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	var sent atomic.Int32
	var lastBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/token":
			_ = r.ParseForm()
			if r.Form.Get("grant_type") != "urn:ietf:params:oauth:grant-type:jwt-bearer" || r.Form.Get("assertion") == "" {
				w.WriteHeader(400)
				return
			}
			_, _ = io.WriteString(w, `{"access_token":"ya29.test","expires_in":3600}`)
		case strings.HasSuffix(r.URL.Path, "/messages:send"):
			if r.Header.Get("Authorization") != "Bearer ya29.test" {
				w.WriteHeader(401)
				return
			}
			_ = json.NewDecoder(r.Body).Decode(&lastBody)
			msg := lastBody["message"].(map[string]any)
			if msg["token"] == "dead-token-0000000" {
				w.WriteHeader(404)
				_, _ = io.WriteString(w, `{"error":{"status":"NOT_FOUND","details":[{"errorCode":"UNREGISTERED"}]}}`)
				return
			}
			sent.Add(1)
			_, _ = io.WriteString(w, `{"name":"projects/p/messages/1"}`)
		}
	}))
	defer srv.Close()
	creds, _ := json.Marshal(map[string]string{"project_id": "p", "client_email": "push@p.iam.gserviceaccount.com",
		"private_key": pemOf(t, "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(key)), "token_uri": srv.URL + "/token"})
	f, err := newFCM(creds, srv.Client())
	if err != nil {
		t.Fatal(err)
	}
	f.api = srv.URL
	n := Note{Title: "Ann", Body: "New message", Tag: "conv:1", Data: map[string]string{"kind": "dm"}}
	if err := f.send(context.Background(), "live-token-0000000", n); err != nil {
		t.Fatalf("send: %v", err)
	}
	if sent.Load() != 1 {
		t.Fatal("nothing reached FCM")
	}
	android := lastBody["message"].(map[string]any)["android"].(map[string]any)["notification"].(map[string]any)
	if android["channel_id"] != channelID || android["visibility"] != "PRIVATE" || android["tag"] != "conv:1" {
		t.Fatalf("android notification %v", android)
	}
	if err := f.send(context.Background(), "dead-token-0000000", n); err != errGone {
		t.Fatalf("a removed app must be reported gone, got %v", err)
	}
}

func TestAPNsSendsWithProviderTokenAndDropsDeadAddresses(t *testing.T) {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	der, _ := x509.MarshalPKCS8PrivateKey(key)
	var got http.Header
	var payload map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/gone") {
			w.WriteHeader(410)
			_, _ = io.WriteString(w, `{"reason":"Unregistered"}`)
			return
		}
		got = r.Header.Clone()
		_ = json.NewDecoder(r.Body).Decode(&payload)
	}))
	defer srv.Close()
	a, err := newAPNs([]byte(pemOf(t, "PRIVATE KEY", der)), "KEY123", "TEAM123", "com.newgen.hub", true, srv.Client())
	if err != nil {
		t.Fatal(err)
	}
	a.host = srv.URL
	n := Note{Title: "Incoming call", Body: "Ann", Tag: "call:1", Urgent: true, Data: map[string]string{"kind": "call"}}
	if err := a.send(context.Background(), "abc", n); err != nil {
		t.Fatalf("send: %v", err)
	}
	if got.Get("apns-topic") != "com.newgen.hub" || !strings.HasPrefix(got.Get("authorization"), "bearer ") || got.Get("apns-push-type") != "alert" {
		t.Fatalf("headers %v", got)
	}
	if payload["kind"] != "call" || payload["aps"].(map[string]any)["thread-id"] != "call:1" {
		t.Fatalf("payload %v", payload)
	}
	if err := a.send(context.Background(), "gone", n); err != errGone {
		t.Fatalf("a removed app must be reported gone, got %v", err)
	}
}

func TestComposeKeepsDetailsToWhatTheOrganizationAllows(t *testing.T) {
	org, acct := uuid.New(), uuid.New()
	note, _ := json.Marshal(map[string]string{"id": "n1", "kind": "ticket", "text": "Ann assigned you ticket T-12", "ref": "T-12"})
	call, _ := json.Marshal(map[string]string{"call": "c1", "kind": "video", "context": "dm", "from_name": "Ann"})
	for _, detail := range []bool{true, false} {
		s := &Service{details: newDetailCache()}
		s.details.seen[org] = detailEntry{detail, time.Now()}
		j, ok := s.compose(context.Background(), realtime.Event{Type: "notification", Channel: realtime.AcctChannel(acct), Org: org.String(), Data: note})
		if !ok || len(j.accounts) != 1 || j.accounts[0] != acct || j.note.Data["ref"] != "T-12" {
			t.Fatalf("notification job %+v", j)
		}
		if detail != strings.Contains(j.note.Body, "T-12") {
			t.Fatalf("details=%v but body %q", detail, j.note.Body)
		}
		c, ok := s.compose(context.Background(), realtime.Event{Type: "call.incoming", Channel: realtime.AcctChannel(acct), Org: org.String(), Data: call})
		if !ok || !c.note.Urgent || c.note.Body != "Incoming video call" || (c.note.Title == "Ann") != detail {
			t.Fatalf("details=%v call %+v", detail, c.note)
		}
	}
	if wanted(realtime.Event{Type: "message.reaction", Channel: "conv:x"}) || wanted(realtime.Event{Type: "notification", Channel: "org:x"}) {
		t.Fatal("only notifications, calls and new messages are pushed")
	}
}

func TestSecretAcceptsTextBase64OrFile(t *testing.T) {
	t.Setenv("HUB_TEST_SECRET", `{"a":1}`)
	if string(secret("HUB_TEST_SECRET")) != `{"a":1}` {
		t.Fatal("json text")
	}
	t.Setenv("HUB_TEST_SECRET", "eyJhIjoxfQ==")
	if string(secret("HUB_TEST_SECRET")) != `{"a":1}` {
		t.Fatal("base64")
	}
}

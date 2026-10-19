// Package testkit starts the real API against the local database and Redis for integration tests.
package testkit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"hub/internal/app"
	"hub/internal/config"
	"hub/internal/crypto"
	"hub/internal/mail"
	"hub/internal/modules"
	"hub/internal/modules/auth"
	"hub/internal/server"
)

type Env struct {
	T    *testing.T
	Srv  *httptest.Server
	Dev  *mail.Dev
	A    *app.App
	Auth *auth.Service
}

func New(t *testing.T) *Env {
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
	au := auth.New(a)
	go a.Bus.Run(context.Background(), a.Hub.Deliver)
	srv := httptest.NewServer(server.Router(server.Deps{App: a, Auth: au, Dev: dev, Modules: modules.All(a, au)}))
	t.Cleanup(srv.Close)
	return &Env{t, srv, dev, a, au}
}

// Do sends a JSON request as a client with its own address (so per-address rate limits never couple unrelated tests).
func (e *Env) Do(method, path string, body any, token string) (int, map[string]any) {
	e.T.Helper()
	var rd *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	} else {
		rd = bytes.NewReader(nil)
	}
	req, _ := http.NewRequest(method, e.Srv.URL+"/v1"+path, rd)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", fmt.Sprintf("10.%d.%d.%d", uuid.New().ID()%250, uuid.New().ID()%250, uuid.New().ID()%250))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		e.T.Fatal(err)
	}
	defer res.Body.Close()
	var out map[string]any
	_ = json.NewDecoder(res.Body).Decode(&out)
	return res.StatusCode, out
}

func (e *Env) MustDo(want int, method, path string, body any, token string) map[string]any {
	e.T.Helper()
	code, out := e.Do(method, path, body, token)
	if code != want {
		e.T.Fatalf("%s %s: want %d got %d: %v", method, path, want, code, out)
	}
	return out
}

func Uniq() string { return "u" + strings.ReplaceAll(uuid.NewString(), "-", "")[:12] + "@example.com" }

const Password = "Passw0rdOK"

// User signs up and signs in; org is optional (the web name to enter).
type User struct {
	Email, Access, Refresh string
	Account                string
}

func (e *Env) Signup(name, email string) {
	e.T.Helper()
	e.MustDo(201, "POST", "/auth/signup", map[string]any{"name": name, "email": email, "password": Password, "quiz": []map[string]string{
		{"question": "pet", "answer": "Rex"}, {"question": "town", "answer": "Nairobi"}, {"question": "school", "answer": "Hub Primary"}}}, "")
}

func (e *Env) Login(email, org string) *User {
	e.T.Helper()
	b := map[string]any{"email": email, "password": Password}
	if org != "" {
		b["org"] = org
	}
	out := e.MustDo(200, "POST", "/auth/login", b, "")
	return &User{Email: email, Access: out["access_token"].(string), Refresh: out["refresh_token"].(string)}
}

// NewOrg signs up an owner, creates an organization and returns the owner signed in to it.
func (e *Env) NewOrg(ownerName string) (*User, string) {
	e.T.Helper()
	email := Uniq()
	e.Signup(ownerName, email)
	u := e.Login(email, "")
	slug := "t-" + strings.ToLower(strings.ReplaceAll(uuid.NewString(), "-", "")[:14])
	e.MustDo(201, "POST", "/orgs", map[string]any{"name": ownerName + " Co", "slug": slug}, u.Access)
	return e.Login(email, slug), slug
}

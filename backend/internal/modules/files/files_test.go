package files_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"io"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/e2ee"
	"hub/internal/orgsettings"
	"hub/internal/testkit"
)

func raw(t *testing.T, e *testkit.Env, method, path string, body []byte, token string) (int, []byte) {
	t.Helper()
	req, _ := http.NewRequest(method, e.Srv.URL+"/v1"+path, bytes.NewReader(body))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("X-Forwarded-For", "10.7."+uuid.NewString()[:2]+".9")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	return res.StatusCode, b
}

func TestEncryptedFilesAndSecureShare(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	org := e.OrgID(slug)
	o := e.Owner(ownerU)
	a := e.Join(ownerU, slug, "Amy", "staff")
	out := e.Join(ownerU, slug, "Outsider", "staff")
	for _, m := range []*testkit.Member{o, a, out} {
		m.Register(e)
	}
	// a folder for Olivia and Amy only
	fk, _ := e2ee.NewConversationKey()
	folder := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "folder", "name": "Finance", "members": []string{a.Person}, "keys": o.WrapFor(e, fk, 1, o.Person, a.Person)}, ownerU.Access)["id"].(string)

	// the file bytes are encrypted by the client; here random bytes stand in for ciphertext
	content := make([]byte, 300_000)
	_, _ = rand.Read(content)
	fid := uuid.Must(uuid.NewV7()).String()
	e.MustDo(201, "POST", "/folders/"+folder+"/files", map[string]any{"id": fid, "epoch": 1, "device": a.DevID, "envelope": a.Seal(org, folder, fid, 1, fk, `{"name":"budget.xlsx"}`)}, a.Access)
	if code, _ := raw(t, e, "PUT", "/files/"+fid+"/content", content, out.Access); code != 404 {
		t.Fatalf("someone else uploaded content: %d", code)
	}
	if code, _ := raw(t, e, "PUT", "/files/"+fid+"/content", content, a.Access); code != 200 {
		t.Fatalf("upload: %d", code)
	}
	if code, _ := raw(t, e, "PUT", "/files/"+fid+"/content", content, a.Access); code != 404 {
		t.Fatalf("content replaced after upload: %d", code)
	}
	if code, got := raw(t, e, "GET", "/files/"+fid+"/content", nil, o.Access); code != 200 || !bytes.Equal(got, content) {
		t.Fatalf("member download: %d", code)
	}
	if code, _ := raw(t, e, "GET", "/files/"+fid+"/content", nil, out.Access); code != 404 {
		t.Fatalf("non-member downloaded: %d", code)
	}
	if code, _ := e.Do("GET", "/folders/"+folder+"/files", nil, out.Access); code != 404 {
		t.Fatalf("non-member listed folder: %d", code)
	}
	if l := e.MustDo(200, "GET", "/folders/"+folder+"/files", nil, a.Access)["files"].([]any); len(l) != 1 {
		t.Fatalf("list: %v", l)
	}

	// a password-protected link allowed twice
	keyEnv := a.Seal(org, folder, fid, 1, fk, `link-key-package`)
	res := e.MustDo(201, "POST", "/files/"+fid+"/shares", map[string]any{"expires_in_hours": 24, "max_downloads": 2, "password": "s3cret-pass", "key_envelope": keyEnv}, a.Access)
	token := res["share"].(map[string]any)["token"].(string)
	if res["approval"] == nil {
		// f3 has one level by default
		t.Fatalf("outside sharing should need approval: %v", res)
	}
	if code, _ := e.Do("GET", "/share/"+token+"/content", nil, ""); code != 404 {
		t.Fatalf("link worked before approval: %d", code)
	}
	e.MustDo(200, "POST", "/approvals/"+res["approval"].(map[string]any)["id"].(string)+"/approve", map[string]any{}, ownerU.Access)
	meta := e.MustDo(200, "GET", "/share/"+token, nil, "")
	if meta["password_protected"] != true || meta["key_envelope"] != nil {
		t.Fatalf("link meta leaks the key package before the password: %v", meta)
	}
	if code, _ := e.Do("POST", "/share/"+token+"/unlock", map[string]any{"password": "wrong"}, ""); code != 401 {
		t.Fatalf("wrong password: %d", code)
	}
	un := e.MustDo(200, "POST", "/share/"+token+"/unlock", map[string]any{"password": "s3cret-pass"}, "")
	pass := un["download_pass"].(string)
	if code, got := raw(t, e, "GET", "/share/"+token+"/content?pass="+pass, nil, ""); code != 200 || !bytes.Equal(got, content) {
		t.Fatalf("link download: %d", code)
	}
	if code, _ := raw(t, e, "GET", "/share/"+token+"/content?pass="+pass, nil, ""); code != 401 {
		t.Fatalf("download pass reused: %d", code)
	}
	if code, _ := raw(t, e, "GET", "/share/"+token+"/content", nil, ""); code != 401 {
		t.Fatalf("no pass for a protected link: %d", code)
	}
	un = e.MustDo(200, "POST", "/share/"+token+"/unlock", map[string]any{"password": "s3cret-pass"}, "")
	if code, _ := raw(t, e, "GET", "/share/"+token+"/content?pass="+un["download_pass"].(string), nil, ""); code != 200 {
		t.Fatalf("second download: %d", code)
	}
	if code, _ := e.Do("POST", "/share/"+token+"/unlock", map[string]any{"password": "s3cret-pass"}, ""); code != 404 {
		t.Fatalf("a used-up link still unlocks: %d", code)
	}

	// turning off outside sharing stops new links
	if err := e.A.DB.WithOrg(context.Background(), uuid.MustParse(org), func(tx pgx.Tx) error {
		s := orgsettings.DefaultSecurity()
		s.ExternalSharing = false
		return orgsettings.Set(context.Background(), e.A, tx, uuid.MustParse(org), "security", s)
	}); err != nil {
		t.Fatal(err)
	}
	if code, _ := e.Do("POST", "/files/"+fid+"/shares", map[string]any{"expires_in_hours": 1, "max_downloads": 1, "key_envelope": keyEnv}, a.Access); code != 403 {
		t.Fatalf("link made while sharing is off: %d", code)
	}
	// deleting the file erases the bytes
	e.MustDo(204, "DELETE", "/files/"+fid, nil, a.Access)
	if code, _ := raw(t, e, "GET", "/files/"+fid+"/content", nil, o.Access); code != 404 {
		t.Fatalf("deleted file still served: %d", code)
	}
}

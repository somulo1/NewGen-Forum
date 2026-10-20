package files_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/e2ee"
	"hub/internal/modules/approvals"
	"hub/internal/modules/files"
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

// A person's uploads stop at their storage limit: the organization's default, or their own limit when they have one.
func TestStorageQuota(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	org := e.OrgID(slug)
	o := e.Owner(ownerU)
	a := e.Join(ownerU, slug, "Amy", "staff")
	for _, m := range []*testkit.Member{o, a} {
		m.Register(e)
	}
	fk, _ := e2ee.NewConversationKey()
	folder := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "folder", "name": "Team", "members": []string{a.Person}, "keys": o.WrapFor(e, fk, 1, o.Person, a.Person)}, ownerU.Access)["id"].(string)
	upload := func(size int) int {
		fid := uuid.Must(uuid.NewV7()).String()
		if code, _ := e.Do("POST", "/folders/"+folder+"/files", map[string]any{"id": fid, "epoch": 1, "device": a.DevID, "envelope": a.Seal(org, folder, fid, 1, fk, `{"name":"x"}`)}, a.Access); code != 201 {
			return code
		}
		code, _ := raw(t, e, "PUT", "/files/"+fid+"/content", make([]byte, size), a.Access)
		return code
	}
	// the default is 5 GB, and Amy uses nothing yet
	if u := e.MustDo(200, "GET", "/storage/usage", nil, a.Access); u["quota_mb"].(float64) != 5120 || u["used_bytes"].(float64) != 0 {
		t.Fatalf("usage: %v", u)
	}
	// staff cannot change limits
	if code, _ := e.Do("PUT", "/storage/people/"+a.Person, map[string]any{"quota_mb": 0}, a.Access); code != 403 {
		t.Fatalf("staff set a limit: %d", code)
	}
	// Amy gets 1 MB of her own: 700 KB fits, another 700 KB does not
	e.MustDo(204, "PUT", "/storage/people/"+a.Person, map[string]any{"quota_mb": 1}, ownerU.Access)
	if c := upload(700 << 10); c != 200 {
		t.Fatalf("first upload: %d", c)
	}
	if c := upload(700 << 10); c != 413 {
		t.Fatalf("over the limit: %d", c)
	}
	if u := e.MustDo(200, "GET", "/storage/usage", nil, a.Access); u["used_bytes"].(float64) != 700<<10 || u["custom"] != true {
		t.Fatalf("usage after: %v", u)
	}
	// her biggest files are listed so she can pick what to delete; deleting one frees its space
	mine := e.MustDo(200, "GET", "/storage/files", nil, a.Access)["files"].([]any)
	if len(mine) != 1 || mine[0].(map[string]any)["size"].(float64) != 700<<10 || mine[0].(map[string]any)["envelope"] == "" {
		t.Fatalf("my files: %v", mine)
	}
	e.MustDo(204, "DELETE", "/files/"+mine[0].(map[string]any)["id"].(string), nil, a.Access)
	if u := e.MustDo(200, "GET", "/storage/usage", nil, a.Access); u["used_bytes"].(float64) != 0 {
		t.Fatalf("usage after deleting: %v", u)
	}
	if c := upload(700 << 10); c != 200 {
		t.Fatalf("after freeing space: %d", c)
	}
	// the owner sees her in the storage list
	found := false
	for _, x := range e.MustDo(200, "GET", "/storage/people", nil, ownerU.Access)["people"].([]any) {
		found = found || x.(map[string]any)["person"] == a.Person
	}
	if !found {
		t.Fatal("Amy is not in the storage list")
	}
	// back on the default (raised to the 100 MB choice), the upload fits
	e.MustDo(204, "PUT", "/storage/people/"+a.Person, map[string]any{"quota_mb": nil}, ownerU.Access)
	sec := e.MustDo(200, "GET", "/settings", nil, ownerU.Access)["security"].(map[string]any)
	sec["userQuotaMB"] = "100"
	e.MustDo(204, "PUT", "/settings/security", sec, ownerU.Access)
	if c := upload(700 << 10); c != 200 {
		t.Fatalf("on the default: %d", c)
	}
	sec["userQuotaMB"] = "7"
	if code, _ := e.Do("PUT", "/settings/security", sec, ownerU.Access); code != 400 {
		t.Fatalf("a limit not on the list: %d", code)
	}
}

// Deleting a file erases its contents from storage; contents that could not be erased at once are retried by the worker.
func TestDeleteErasesContents(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	org := e.OrgID(slug)
	o := e.Owner(ownerU)
	o.Register(e)
	fk, _ := e2ee.NewConversationKey()
	folder := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "folder", "name": "Mine", "keys": o.WrapFor(e, fk, 1, o.Person)}, ownerU.Access)["id"].(string)
	fid := uuid.Must(uuid.NewV7()).String()
	e.MustDo(201, "POST", "/folders/"+folder+"/files", map[string]any{"id": fid, "epoch": 1, "device": o.DevID, "envelope": o.Seal(org, folder, fid, 1, fk, `{"name":"x"}`)}, ownerU.Access)
	if code, _ := raw(t, e, "PUT", "/files/"+fid+"/content", make([]byte, 4096), ownerU.Access); code != 200 {
		t.Fatalf("upload: %d", code)
	}
	onDisk := filepath.Join(e.A.Cfg.BlobDir, "o", org, fid)
	if _, err := os.Stat(onDisk); err != nil {
		t.Fatalf("contents not stored: %v", err)
	}
	e.MustDo(204, "DELETE", "/files/"+fid, nil, ownerU.Access)
	if _, err := os.Stat(onDisk); !os.IsNotExist(err) {
		t.Fatal("contents still on disk after delete")
	}

	// contents left waiting (the file server was unreachable at delete time) are erased by the worker's sweep
	left := filepath.Join(e.A.Cfg.BlobDir, "o", org, "leftover-"+uuid.NewString())
	_ = os.WriteFile(left, []byte("old"), 0o600)
	ref, _ := filepath.Rel(e.A.Cfg.BlobDir, left)
	if err := e.A.DB.Tx(context.Background(), func(tx pgx.Tx) error {
		_, err := tx.Exec(context.Background(), `insert into hub.blob_purge (org_id, ref) values ($1,$2)`, org, ref)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	svc := files.New(e.A, approvals.New(e.A))
	if done, _, err := svc.PurgeDue(context.Background(), 500); err != nil || done < 1 {
		t.Fatalf("purge: %d %v", done, err)
	}
	if _, err := os.Stat(left); !os.IsNotExist(err) {
		t.Fatal("waiting contents were not erased")
	}
	var n int
	_ = e.A.DB.Tx(context.Background(), func(tx pgx.Tx) error {
		return tx.QueryRow(context.Background(), `select count(*) from hub.blob_purge where ref=$1`, ref).Scan(&n)
	})
	if n != 0 {
		t.Fatal("the erased reference is still queued")
	}
}

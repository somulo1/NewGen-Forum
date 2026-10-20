package blob_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"

	"hub/internal/blob"
)

// fakeS3 keeps objects in memory and refuses unsigned requests.
type fakeS3 struct {
	mu   sync.Mutex
	objs map[string][]byte
	down bool
}

func (f *fakeS3) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.down {
		w.WriteHeader(503)
		return
	}
	if !strings.HasPrefix(r.Header.Get("Authorization"), "AWS4-HMAC-SHA256 Credential=KEY/") || r.Header.Get("x-amz-content-sha256") == "" {
		w.WriteHeader(403)
		return
	}
	k := r.URL.Path
	switch r.Method {
	case http.MethodPut:
		b, _ := io.ReadAll(r.Body)
		f.objs[k] = b
	case http.MethodHead:
		if k == "/bucket" {
			return
		}
		b, ok := f.objs[k]
		if !ok {
			w.WriteHeader(404)
			return
		}
		w.Header().Set("Content-Length", strconv.Itoa(len(b)))
	case http.MethodGet:
		b, ok := f.objs[k]
		if !ok {
			w.WriteHeader(404)
			return
		}
		from := 0
		if rg := r.Header.Get("Range"); rg != "" {
			from, _ = strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(rg, "bytes="), "-"))
		}
		w.WriteHeader(206)
		_, _ = w.Write(b[from:])
	case http.MethodDelete:
		delete(f.objs, k)
		w.WriteHeader(204)
	}
}

func TestPoolUsesFileServerFallsBackAndErases(t *testing.T) {
	ctx := context.Background()
	fs := &fakeS3{objs: map[string][]byte{}}
	srv := httptest.NewServer(fs)
	defer srv.Close()
	dir := t.TempDir()
	pool, err := blob.NewPool(dir, []string{"s3://KEY:SECRET@" + strings.TrimPrefix(srv.URL, "http://") + "/bucket?tls=0"})
	if err != nil {
		t.Fatal(err)
	}
	data := bytes.Repeat([]byte("ciphertext"), 5000)

	// the file server is up: the bytes go there, and the reference says so
	ref, n, _, err := pool.Put(ctx, "o/org1/file1", bytes.NewReader(data), int64(len(data)))
	if err != nil || n != int64(len(data)) || ref != "fs1:o/org1/file1" {
		t.Fatalf("put: %q %d %v", ref, n, err)
	}
	if _, ok := fs.objs["/bucket/o/org1/file1"]; !ok {
		t.Fatal("not on the file server")
	}
	r, size, err := pool.Get(ctx, ref)
	if err != nil || size != int64(len(data)) {
		t.Fatalf("get: %d %v", size, err)
	}
	_, _ = r.Seek(10, io.SeekStart)
	got, _ := io.ReadAll(r)
	r.Close()
	if !bytes.Equal(got, data[10:]) {
		t.Fatal("ranged read differs")
	}
	// deleting removes it from the file server
	if err := pool.Delete(ctx, ref); err != nil {
		t.Fatal(err)
	}
	if _, ok := fs.objs["/bucket/o/org1/file1"]; ok {
		t.Fatal("still on the file server after delete")
	}

	// the file server goes down: new files land on this server's disk, and deleting erases them there
	fs.down = true
	ref2, _, _, err := pool.Put(ctx, "o/org1/file2", bytes.NewReader(data), int64(len(data)))
	if err != nil || ref2 != "o/org1/file2" {
		t.Fatalf("fallback put: %q %v", ref2, err)
	}
	if _, err := os.Stat(filepath.Join(dir, "o/org1/file2")); err != nil {
		t.Fatal("not on local disk")
	}
	if err := pool.Delete(ctx, ref2); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "o/org1/file2")); !os.IsNotExist(err) {
		t.Fatal("still on local disk after delete")
	}
	// nothing is left in the spool either
	if left, _ := os.ReadDir(filepath.Join(dir, ".spool")); len(left) != 0 {
		t.Fatalf("spool not cleaned: %d", len(left))
	}
	// a delete on a server that is down fails, so the caller can queue it for a retry
	ref3 := "fs1:o/org1/file3"
	if err := pool.Delete(ctx, ref3); err == nil {
		t.Fatal("a delete on a down server reported success")
	}
	// too large is refused
	if _, _, _, err := pool.Put(ctx, "o/org1/file4", bytes.NewReader(data), 10); err != blob.ErrTooLarge {
		t.Fatalf("too large: %v", err)
	}
}

package roles_test

import (
	"bytes"
	"encoding/base64"
	"io"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"hub/internal/testkit"
)

func getRaw(t *testing.T, e *testkit.Env, path, token string) (int, string, []byte) {
	t.Helper()
	req, _ := http.NewRequest("GET", e.Srv.URL+"/v1"+path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Forwarded-For", "10.9."+uuid.NewString()[:2]+".9")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	return res.StatusCode, res.Header.Get("Content-Type"), b
}

// Each person's own picture is served to colleagues, so a post or comment shows its author's face, not the viewer's.
func TestEveryonesPhotoIsServedAndFlagged(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	a := e.Join(ownerU, slug, "Ann", "staff")
	b := e.Join(ownerU, slug, "Bea", "staff")
	pngA := []byte("\x89PNG....ann")
	pngB := []byte("\x89PNG....bea!!")
	enc := func(x []byte) string { return "data:image/png;base64," + base64.StdEncoding.EncodeToString(x) }
	e.MustDo(204, "PATCH", "/me", map[string]any{"photo": enc(pngA)}, a.Access)

	// Bea sees that Ann has a photo and Bea has none yet
	flags := map[string]any{}
	for _, p := range e.MustDo(200, "GET", "/people?limit=100", nil, b.Access)["people"].([]any) {
		m := p.(map[string]any)
		flags[m["id"].(string)] = m["has_photo"]
	}
	if flags[a.Person] != true || flags[b.Person] != false {
		t.Fatalf("has_photo wrong: %v", flags)
	}
	// Bea fetches Ann's picture and gets Ann's bytes, as a picture
	code, typ, body := getRaw(t, e, "/people/"+a.Person+"/photo", b.Access)
	if code != 200 || typ != "image/png" || !bytes.Equal(body, pngA) {
		t.Fatalf("Ann's photo: %d %s %q", code, typ, body)
	}
	// Bea has none: not found, never someone else's
	if code, _, _ := getRaw(t, e, "/people/"+b.Person+"/photo", a.Access); code != 404 {
		t.Fatalf("a person without a photo should be 404, got %d", code)
	}
	// once Bea sets hers, each person's URL gives that person's own picture
	e.MustDo(204, "PATCH", "/me", map[string]any{"photo": enc(pngB)}, b.Access)
	_, _, ba := getRaw(t, e, "/people/"+b.Person+"/photo", a.Access)
	_, _, aa := getRaw(t, e, "/people/"+a.Person+"/photo", a.Access)
	if !bytes.Equal(ba, pngB) || !bytes.Equal(aa, pngA) {
		t.Fatal("the pictures were mixed up")
	}
}

package chat_test

import (
	"bytes"
	"testing"

	"github.com/google/uuid"

	"hub/internal/testkit"
)

// Notes are held by the server (encrypted at rest with the organization's key), not end-to-end: no device key is needed to make
// a notes vault, write a note, share it or read it, a shared note opens for the people it was shared with, and nobody else.
func TestNotesAreHeldByTheServer(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	a := e.Join(ownerU, slug, "Amy", "staff")
	b := e.Join(ownerU, slug, "Ben", "staff")
	c := e.Join(ownerU, slug, "Cy", "staff")
	_ = slug
	vault := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "note"}, a.Access)["id"].(string) // no keys at all
	id := uuid.Must(uuid.NewV7()).String()
	body := map[string]any{"title": "Plan", "body": "<p>secret plan text</p>"}
	e.MustDo(201, "POST", "/containers/"+vault+"/records", map[string]any{"id": id, "collection": "note", "body": body, "meta": map[string]any{}}, a.Access)
	got := func(user string, conv string) map[string]any {
		t.Helper()
		recs := e.MustDo(200, "GET", "/containers/"+conv+"/records?collection=note", nil, user)["records"].([]any)
		if len(recs) != 1 {
			t.Fatalf("%d records", len(recs))
		}
		return recs[0].(map[string]any)
	}
	r := got(a.Access, vault)
	if r["body"].(map[string]any)["title"] != "Plan" || r["envelope"] != nil {
		t.Fatalf("the note did not come back as written: %v", r)
	}
	var stored []byte
	e.RootScan(`select ciphertext from hub.records where id=$1`, []any{id}, &stored)
	if bytes.Contains(stored, []byte("secret plan text")) || !bytes.HasPrefix(stored, []byte("HUBP1:")) {
		t.Fatalf("the note is not encrypted at rest")
	}
	e.MustDo(200, "PUT", "/containers/"+vault+"/records/"+id, map[string]any{"id": id, "collection": "note", "body": map[string]any{"title": "Plan 2", "body": "<p>x</p>"}, "meta": map[string]any{}}, a.Access)
	if got(a.Access, vault)["body"].(map[string]any)["title"] != "Plan 2" {
		t.Fatal("an edit was not kept")
	}
	if code, _ := e.Do("GET", "/containers/"+vault+"/records", nil, b.Access); code != 404 {
		t.Fatalf("someone else read a private notes vault: %d", code)
	}
	shared := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "sharednote"}, a.Access)["id"].(string)
	sid := uuid.Must(uuid.NewV7()).String()
	e.MustDo(201, "POST", "/containers/"+shared+"/records", map[string]any{"id": sid, "collection": "note", "body": body, "meta": map[string]any{}}, a.Access)
	e.MustDo(200, "POST", "/conversations/"+shared+"/members", map[string]any{"people": []string{b.Person}}, a.Access)
	if got(b.Access, shared)["body"].(map[string]any)["title"] != "Plan" {
		t.Fatal("the person it was shared with cannot read it")
	}
	if code, _ := e.Do("PUT", "/containers/"+shared+"/records/"+sid, map[string]any{"id": sid, "collection": "note", "body": body, "meta": map[string]any{}}, b.Access); code != 403 {
		t.Fatalf("a read-only person edited a shared note: %d", code)
	}
	if code, _ := e.Do("GET", "/containers/"+shared+"/records", nil, c.Access); code != 404 {
		t.Fatalf("a stranger read a shared note: %d", code)
	}
}

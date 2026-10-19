package chat_test

import (
	"testing"

	"github.com/google/uuid"

	"hub/internal/e2ee"
	"hub/internal/testkit"
)

// A shared note is a container with a note in it: the owner and anyone given edit rights can change it, everyone else added
// can only read it, and only the owner decides who is in.
func TestSharedNoteReadOnlyAndEditRights(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	org := e.OrgID(slug)
	o := e.Owner(ownerU)
	ed := e.Join(ownerU, slug, "Edie", "")
	rd := e.Join(ownerU, slug, "Rae", "")
	for _, m := range []*testkit.Member{o, ed, rd} {
		m.Register(e)
	}
	key, _ := e2ee.NewConversationKey()
	conv := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "sharednote", "members": []string{ed.Person, rd.Person}, "keys": o.WrapFor(e, key, 1, o.Person, ed.Person, rd.Person)}, ownerU.Access)["id"].(string)

	save := func(m *testkit.Member, method, id, text string, want int) {
		t.Helper()
		path := "/containers/" + conv + "/records"
		if method == "PUT" {
			path += "/" + id
		}
		e.MustDo(want, method, path, map[string]any{"id": id, "collection": "note", "epoch": 1, "device": m.DevID, "envelope": m.Seal(org, conv, id, 1, key, text), "meta": map[string]any{}}, m.Access)
	}
	note := uuid.Must(uuid.NewV7()).String()
	save(rd, "POST", uuid.Must(uuid.NewV7()).String(), "reader makes one", 403)
	save(o, "POST", note, `{"title":"Plan","body":"v1"}`, 201)
	// nobody but the owner has edit rights yet
	save(ed, "PUT", note, "edit before rights", 403)
	// the owner gives Edie edit rights; Rae stays read only
	e.MustDo(204, "PATCH", "/conversations/"+conv+"/members/"+ed.Person, map[string]any{"role": "admin"}, ownerU.Access)
	save(ed, "PUT", note, `{"title":"Plan","body":"v2"}`, 200)
	save(rd, "PUT", note, "reader edits", 403)
	if code, _ := e.Do("DELETE", "/containers/"+conv+"/records/"+note, nil, rd.Access); code != 403 {
		t.Fatalf("a reader deleted the note: %d", code)
	}
	// only the owner changes rights and members
	if code, _ := e.Do("PATCH", "/conversations/"+conv+"/members/"+rd.Person, map[string]any{"role": "admin"}, ed.Access); code != 403 {
		t.Fatalf("an editor gave rights: %d", code)
	}
	if code, _ := e.Do("DELETE", "/conversations/"+conv+"/members/"+rd.Person, map[string]any{"device": ed.DevID, "keys": []any{}}, ed.Access); code != 403 {
		t.Fatalf("an editor removed someone: %d", code)
	}
	// edit rights taken away again
	e.MustDo(204, "PATCH", "/conversations/"+conv+"/members/"+ed.Person, map[string]any{"role": "member"}, ownerU.Access)
	save(ed, "PUT", note, "after rights removed", 403)
	// the owner can always change or delete it
	save(o, "PUT", note, `{"title":"Plan","body":"v3"}`, 200)
	if code, _ := e.Do("DELETE", "/containers/"+conv+"/records/"+note, nil, ownerU.Access); code != 204 {
		t.Fatalf("owner could not delete: %d", code)
	}
}

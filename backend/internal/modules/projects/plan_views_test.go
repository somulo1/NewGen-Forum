package projects_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"

	"hub/internal/e2ee"
	"hub/internal/testkit"
)

// View own projects shows the person's teams; View all projects shows every project, to look at but not to change.
func TestViewOwnAndViewAllProjects(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	o := e.Owner(ownerU)
	amy := e.Join(ownerU, slug, "Amy", "hod")
	ben := e.Join(ownerU, slug, "Ben", "hod")
	for _, m := range []*testkit.Member{o, amy, ben} {
		m.Register(e)
	}
	e.MustDo(204, "PUT", "/approvals/flows/f2", map[string]any{"levels": 0}, ownerU.Access)
	e.MustDo(204, "PUT", "/approvals/flows/f8", map[string]any{"levels": 0}, ownerU.Access)
	key, _ := e2ee.NewConversationKey()
	conv := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "project", "name": "Secret plan", "about": "why", "members": []string{amy.Person}, "keys": o.WrapFor(e, key, 1, o.Person, amy.Person)}, ownerU.Access)["id"].(string)
	e.MustDo(201, "POST", "/projects", map[string]any{"conversation": conv, "start": "2026-01-01", "due": "2027-01-01"}, ownerU.Access)
	task := uuid.Must(uuid.NewV7()).String()
	e.MustDo(200, "PUT", "/projects/"+conv+"/items/"+task, map[string]any{"collection": "task", "data": map[string]any{"title": "T"}}, ownerU.Access)
	list := func(m *testkit.Member) []any {
		return e.MustDo(200, "GET", "/projects", nil, m.Access)["projects"].([]any)
	}
	if l := list(amy); len(l) != 1 || l[0].(map[string]any)["member"] != true {
		t.Fatalf("own: %v", l)
	}
	if len(list(ben)) != 0 {
		t.Fatal("someone off the team sees it")
	}
	if c, _ := e.Do("GET", "/projects/"+conv+"/plan", nil, ben.Access); c != 404 {
		t.Fatalf("plan read: %d", c)
	}
	right := func(m *testkit.Member, name string, on bool) {
		e.MustDo(200, "PUT", "/people/"+m.Person+"/rights", map[string]any{"module": "projects", "right": name, "allowed": on}, ownerU.Access)
	}
	raw := func(method, path string, body []byte, token string) int {
		req, _ := http.NewRequest(method, e.Srv.URL+"/v1"+path, bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		_, _ = io.Copy(io.Discard, res.Body)
		return res.StatusCode
	}
	fid := uuid.Must(uuid.NewV7()).String()
	e.MustDo(201, "POST", "/folders/"+conv+"/files", map[string]any{"id": fid, "epoch": 1, "device": o.DevID, "envelope": o.Seal(e.OrgID(slug), conv, fid, 1, key, `{"name":"a.txt"}`)}, ownerU.Access)
	if c := raw("PUT", "/files/"+fid+"/content", make([]byte, 2000), ownerU.Access); c != 200 {
		t.Fatalf("upload: %d", c)
	}
	if c := raw("GET", "/files/"+fid+"/content", nil, ben.Access); c != 404 {
		t.Fatalf("file read off the team: %d", c)
	}
	right(ben, "View all projects", true)
	if c := raw("GET", "/files/"+fid+"/content", nil, ben.Access); c != 200 {
		t.Fatalf("a viewer cannot open project files: %d", c)
	}
	l := list(ben)
	if len(l) != 1 || l[0].(map[string]any)["member"] != false || l[0].(map[string]any)["name"] != "Secret plan" || len(l[0].(map[string]any)["team_names"].([]any)) != 2 {
		t.Fatalf("all: %v", l)
	}
	if n := len(e.MustDo(200, "GET", "/projects/"+conv+"/plan", nil, ben.Access)["items"].([]any)); n != 1 {
		t.Fatalf("plan items: %d", n)
	}
	if c, _ := e.Do("PUT", "/projects/"+conv+"/items/"+task, map[string]any{"collection": "task", "data": map[string]any{"title": "Changed"}}, ben.Access); c != 404 {
		t.Fatalf("a viewer changed a task: %d", c)
	}
	if c, _ := e.Do("PATCH", "/projects/"+conv, map[string]any{"status": "Done"}, ben.Access); c != 404 {
		t.Fatalf("a viewer changed the status: %d", c)
	}
	right(amy, "View own projects", false)
	if len(list(amy)) != 0 {
		t.Fatal("own projects still listed without the right")
	}
}

// An organization made before these rights gets them on its next sign-in, and each role keeps what it could do.
func TestOlderOrganizationGetsTheNewRights(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	amy := e.Join(ownerU, slug, "Amy", "hod")
	e.RemoveRights(slug, "wall", "Hide posts", "Delete others' posts")
	e.RemoveRights(slug, "projects", "View own projects", "View all projects")
	rights := e.MustDo(200, "GET", "/me", nil, amy.Access)
	raw := fmt.Sprint(rights)
	for _, want := range []string{"Hide posts", "View own projects"} {
		if !strings.Contains(raw, want) {
			t.Fatalf("%q missing after upgrade: %s", want, raw[:200])
		}
	}
}

package projects_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"hub/internal/e2ee"
	"hub/internal/testkit"
)

func TestGitHubDrivesTasksWithConsent(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	pm := e.Join(ownerU, slug, "Hana", "hod")
	amy := e.Join(ownerU, slug, "Amy", "hod")
	for _, m := range []*testkit.Member{pm, amy} {
		m.Register(e)
	}
	key, _ := e2ee.NewConversationKey()
	conv := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "project", "name": "Alpha", "members": []string{amy.Person}, "keys": pm.WrapFor(e, key, 1, pm.Person, amy.Person)}, pm.Access)["id"].(string)
	res := e.MustDo(201, "POST", "/projects", map[string]any{"conversation": conv, "start": "2026-09-01", "due": "2026-12-15", "method": "waterfall"}, pm.Access)
	for i := 0; i < 3; i++ {
		e.Do("POST", "/approvals/"+res["approval"].(map[string]any)["id"].(string)+"/approve", map[string]any{}, ownerU.Access)
	}
	proj := res["project"].(map[string]any)
	code := proj["code"].(string)
	var first [4]string // the first status of each kind (0 to 3)
	for _, x := range proj["workflow"].([]any) {
		st := x.(map[string]any)
		if c := int(st["cat"].(float64)); first[c] == "" {
			first[c] = st["key"].(string)
		}
	}
	tid := uuid.Must(uuid.NewV7()).String()
	e.MustDo(200, "PUT", "/projects/"+conv+"/items/"+tid, map[string]any{"collection": "task", "data": map[string]any{"title": "Fix the login timeout", "status": first[0], "assignee": amy.Person, "due": "2026-10-01"}}, pm.Access)
	branch := code + "-1_fix_login_timeout"

	// linking: the owner links and gets the secret once; someone else on the team cannot
	if code, _ := e.Do("PUT", "/projects/"+conv+"/github", map[string]any{"repo": "acme/hub"}, amy.Access); code != 403 && code != 200 {
		t.Fatalf("link by a team member: %d", code)
	}
	g := e.MustDo(200, "PUT", "/projects/"+conv+"/github", map[string]any{"repo": "https://github.com/acme/hub.git", "new_secret": true}, pm.Access)
	secret, hook := g["secret"].(string), g["hook"].(string)
	if g["repo"] != "acme/hub" || secret == "" || hook == "" {
		t.Fatalf("link: %v", g)
	}
	if again := e.MustDo(200, "GET", "/projects/"+conv+"/github", nil, amy.Access); again["secret"] != nil {
		t.Fatal("the secret was shown again")
	}
	deliver := func(event, sign string, payload map[string]any) int {
		b, _ := json.Marshal(payload)
		if sign == "" {
			m := hmac.New(sha256.New, []byte(secret))
			m.Write(b)
			sign = "sha256=" + hex.EncodeToString(m.Sum(nil))
		}
		req, _ := http.NewRequest("POST", e.Srv.URL+hook, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-GitHub-Event", event)
		req.Header.Set("X-GitHub-Delivery", uuid.NewString())
		req.Header.Set("X-Hub-Signature-256", sign)
		r, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		r.Body.Close()
		return r.StatusCode
	}
	made := map[string]any{"ref": branch, "ref_type": "branch", "sender": map[string]any{"login": "amy-dev"}}
	if c := deliver("create", "sha256=00", made); c != 401 {
		t.Fatalf("a delivery with a wrong signature: %d", c)
	}
	if c := deliver("ping", "", map[string]any{"zen": "hi"}); c != 204 {
		t.Fatalf("ping: %d", c)
	}
	if c := deliver("create", "", made); c != 204 {
		t.Fatalf("branch made: %d", c)
	}
	events := func() []any {
		return e.MustDo(200, "GET", "/projects/"+conv+"/github", nil, amy.Access)["events"].([]any)
	}
	ev := events()[0].(map[string]any)
	if ev["task"] != tid || ev["suggest"] != first[1] || ev["state"] != "pending" || ev["kind"] != "branch_created" {
		t.Fatalf("suggestion for the branch: %v", ev)
	}
	// someone on the team accepts it (after moving the task through the usual rules); it cannot be decided twice
	e.MustDo(204, "POST", "/projects/"+conv+"/github/events/"+ev["id"].(string), map[string]any{"decision": "accepted"}, amy.Access)
	if c, _ := e.Do("POST", "/projects/"+conv+"/github/events/"+ev["id"].(string), map[string]any{"decision": "dismissed"}, amy.Access); c != 409 {
		t.Fatalf("decided twice: %d", c)
	}

	// automatic: a pull request whose branch names the task key moves it to "being checked" at once
	e.MustDo(200, "PUT", "/projects/"+conv+"/github", map[string]any{"repo": "acme/hub", "auto": true}, pm.Access)
	pr := map[string]any{"action": "opened", "sender": map[string]any{"login": "amy-dev"}, "pull_request": map[string]any{"title": "Fix login", "html_url": "https://github.com/acme/hub/pull/9", "head": map[string]any{"ref": "feature/" + code + "-1-login"}}}
	if c := deliver("pull_request", "", pr); c != 204 {
		t.Fatalf("pull request: %d", c)
	}
	ev = events()[0].(map[string]any)
	if ev["kind"] != "pr_opened" || ev["state"] != "auto" || ev["suggest"] != first[2] {
		t.Fatalf("automatic move: %v", ev)
	}
	// a branch nobody's task uses is recorded for the project only
	deliver("create", "", map[string]any{"ref": "experiment", "ref_type": "branch"})
	if ev = events()[0].(map[string]any); ev["task"] != nil || ev["state"] != "none" {
		t.Fatalf("unmatched branch: %v", ev)
	}

	// the repository's branches, from GitHub (here a stand-in)
	gh := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/acme/hub/branches" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"name":"main"},{"name":"` + branch + `"}]`))
	}))
	defer gh.Close()
	t.Setenv("HUB_GITHUB_API", gh.URL)
	bs := e.MustDo(200, "GET", "/projects/"+conv+"/github/branches", nil, amy.Access)["branches"].([]any)
	if len(bs) != 2 || bs[1] != branch {
		t.Fatalf("branches: %v", bs)
	}
	// unlinking stops deliveries
	e.MustDo(204, "DELETE", "/projects/"+conv+"/github", nil, pm.Access)
	if c := deliver("create", "", made); c != 404 {
		t.Fatalf("delivery after unlinking: %d", c)
	}
}

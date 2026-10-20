package notify_test

import (
	"strings"
	"testing"
	"time"

	"hub/internal/testkit"
)

func TestTaggedPeopleGetABellAndAnEmail(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	sam := e.Join(ownerU, slug, "Sam", "staff")
	e.MustDo(204, "POST", "/mentions", map[string]any{"where": "ticket", "ref": "C-12", "people": []string{sam.Person}}, ownerU.Access)
	n := e.MustDo(200, "GET", "/notifications", nil, sam.Access)["notifications"].([]any)
	if len(n) == 0 || !strings.Contains(n[0].(map[string]any)["text"].(string), "tagged you") {
		t.Fatalf("no bell: %v", n)
	}
	var got bool
	for i := 0; i < 40 && !got; i++ {
		for _, m := range e.Dev.To(sam.Email) {
			if strings.Contains(m.Subject, "tagged you") && strings.Contains(m.HTML, "You were tagged") {
				got = true
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	if !got {
		t.Fatal("no email to the tagged person")
	}
	e.MustDo(400, "POST", "/mentions", map[string]any{"where": "nowhere", "ref": "x", "people": []string{sam.Person}}, ownerU.Access)
}

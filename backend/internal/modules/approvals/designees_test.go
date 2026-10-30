package approvals_test

import (
	"testing"

	"hub/internal/testkit"
)

func TestSeveralPeopleCanBeChosenForALevel(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	a := e.Join(ownerU, slug, "Amy", "staff")
	b := e.Join(ownerU, slug, "Ben", "staff")
	e.MustDo(204, "PUT", "/approvals/flows/f1", map[string]any{"levels": 2}, ownerU.Access)
	count := func() int {
		for _, f := range e.MustDo(200, "GET", "/approvals/flows", nil, ownerU.Access)["flows"].([]any) {
			m := f.(map[string]any)
			if m["id"] == "f1" {
				return len(m["designees"].([]any))
			}
		}
		return -1
	}
	e.MustDo(204, "PUT", "/approvals/flows/f1/approvers/1", map[string]any{"people": []string{a.Person, b.Person}}, ownerU.Access)
	if n := count(); n != 2 {
		t.Fatalf("two chosen, listed %d", n)
	}
	e.MustDo(204, "PUT", "/approvals/flows/f1/approvers/1", map[string]any{"people": []string{b.Person}}, ownerU.Access)
	if n := count(); n != 1 {
		t.Fatalf("one left, listed %d", n)
	}
	e.MustDo(204, "PUT", "/approvals/flows/f1/approvers/1", map[string]any{"people": []string{}}, ownerU.Access)
	if n := count(); n != 0 {
		t.Fatalf("cleared, listed %d", n)
	}
	if code, _ := e.Do("PUT", "/approvals/flows/f1/approvers/1", map[string]any{"people": []string{a.Person}}, a.Access); code != 403 {
		t.Fatalf("someone without the right chose approvers: %d", code)
	}
	e.MustDo(400, "PUT", "/approvals/flows/f1/approvers/3", map[string]any{"people": []string{a.Person}}, ownerU.Access)
}

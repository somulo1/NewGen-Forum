package roles_test

import (
	"strings"
	"testing"

	"hub/internal/testkit"
)

// A person writes their own professional headline; everyone in the organization sees it next to their name, nobody else can change it.
func TestProfessionalHeadline(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	a := e.Join(ownerU, slug, "Ann", "staff")
	b := e.Join(ownerU, slug, "Bea", "staff")

	e.MustDo(204, "PATCH", "/me", map[string]any{"headline": "  Software   developer | DevOps "}, a.Access)
	if got := e.MustDo(200, "GET", "/me", nil, a.Access)["profile"].(map[string]any)["headline"]; got != "Software developer | DevOps" {
		t.Fatalf("headline not saved tidily: %q", got)
	}
	// a colleague sees it in the directory and on the person's page
	if got := e.MustDo(200, "GET", "/people/"+a.Person, nil, b.Access)["headline"]; got != "Software developer | DevOps" {
		t.Fatalf("colleague does not see it: %v", got)
	}
	found := false
	for _, p := range e.MustDo(200, "GET", "/people?limit=100", nil, b.Access)["people"].([]any) {
		if m := p.(map[string]any); m["id"] == a.Person && m["headline"] == "Software developer | DevOps" {
			found = true
		}
	}
	if !found {
		t.Fatal("the headline is not in the directory list")
	}
	// Bea's own is separate, and a change by Ann leaves it alone
	if got := e.MustDo(200, "GET", "/people/"+b.Person, nil, a.Access)["headline"]; got != "" {
		t.Fatalf("someone else's headline appeared: %v", got)
	}
	// too long is refused; changing and clearing work
	if code, _ := e.Do("PATCH", "/me", map[string]any{"headline": strings.Repeat("x", 121)}, a.Access); code != 400 {
		t.Fatalf("a 121-character headline was accepted: %d", code)
	}
	e.MustDo(204, "PATCH", "/me", map[string]any{"headline": "Team lead"}, a.Access)
	if got := e.MustDo(200, "GET", "/people/"+a.Person, nil, b.Access)["headline"]; got != "Team lead" {
		t.Fatalf("change not saved: %v", got)
	}
	e.MustDo(204, "PATCH", "/me", map[string]any{"headline": ""}, a.Access)
	if got := e.MustDo(200, "GET", "/people/"+a.Person, nil, b.Access)["headline"]; got != "" {
		t.Fatalf("headline not cleared: %v", got)
	}
}

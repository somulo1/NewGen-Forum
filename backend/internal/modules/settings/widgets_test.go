package settings

import "testing"

func TestParseWidgets(t *testing.T) {
	good := map[string]any{"notes": map[string]any{"enabled": true, "items": 5, "managers": []any{"hod"}}, "tickets": map[string]any{"title": " Desk "}}
	got, err := parseWidgets(good)
	if err != nil {
		t.Fatalf("a valid configuration was refused: %v", err)
	}
	if got["tickets"].Title != "Desk" || got["notes"].Items != 5 {
		t.Fatalf("parsed: %+v", got)
	}
	if c, err := parseWidgets(map[string]any{"notes": map[string]any{"enabled": true, "see": "Specific groups"}}); err != nil || !c["notes"].Enabled {
		t.Fatalf("an older page's audience field should be dropped, not refused: %v %v", c, err)
	}
	bad := []map[string]any{
		{"notes": "on"}, // not an object
		{"notes": map[string]any{"enabled": "yes"}},         // switched on is not a yes or no
		{"notes": map[string]any{"items": 50}},              // too many entries on Home
		{"notes": map[string]any{"managers": []any{"A B"}}}, // not a role
		{"Not A Widget!": map[string]any{"enabled": true}},  // not a widget id
	}
	for i, in := range bad {
		if _, err := parseWidgets(in); err == nil {
			t.Fatalf("case %d: an invalid configuration was accepted", i)
		}
	}
}

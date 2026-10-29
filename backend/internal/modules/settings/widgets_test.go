package settings

import "testing"

func TestValidWidgets(t *testing.T) {
	good := map[string]any{"notes": map[string]any{"enabled": true, "see": "Everyone"}, "tickets": map[string]any{"title": "Desk"}}
	if err := validWidgets(good); err != nil {
		t.Fatalf("a valid configuration was refused: %v", err)
	}
	bad := []map[string]any{
		{"notes": "on"}, // not an object
		{"notes": map[string]any{"enabled": "yes"}},        // switched on is not a yes or no
		{"notes": map[string]any{"see": "Nobody"}},         // not a known audience
		{"Not A Widget!": map[string]any{"enabled": true}}, // not a widget id
	}
	for i, in := range bad {
		if err := validWidgets(in); err == nil {
			t.Fatalf("case %d: an invalid configuration was accepted", i)
		}
	}
}

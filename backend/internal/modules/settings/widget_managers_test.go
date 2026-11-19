package settings_test

import (
	"testing"

	"hub/internal/testkit"
)

// A role named as a widget's manager changes that widget's name, switch and Home entries; not who manages it, and not other widgets.
func TestWidgetManagersChangeOnlyTheirWidget(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	sam := e.Join(ownerU, slug, "Sam", "staff")
	e.MustDo(204, "PUT", "/settings/widgets", map[string]any{
		"tickets": map[string]any{"enabled": false, "title": "Tickets", "managers": []string{"staff"}},
		"files":   map[string]any{"enabled": true, "title": "Files"},
	}, ownerU.Access)

	mine := e.MustDo(200, "GET", "/settings/widgets", nil, sam.Access)
	if mine["tickets"] == nil || mine["files"] == nil {
		t.Fatalf("a manager should read the widget they manage (even switched off) and the switched-on ones: %v", mine)
	}
	e.MustDo(204, "PUT", "/settings/widgets", map[string]any{"tickets": map[string]any{"enabled": true, "title": "Help desk", "items": 6, "managers": []string{"staff"}}}, sam.Access)
	refused := []map[string]any{
		{"tickets": map[string]any{"enabled": true, "title": "Help desk", "items": 6, "managers": []string{"staff", "hod"}}},
		{"files": map[string]any{"enabled": false, "title": "Files"}},
	}
	for i, body := range refused {
		if c, _ := e.Do("PUT", "/settings/widgets", body, sam.Access); c != 403 {
			t.Fatalf("case %d: a manager changed what is not theirs: %d", i, c)
		}
	}
	all := e.MustDo(200, "GET", "/settings/widgets", nil, ownerU.Access)
	tk := all["tickets"].(map[string]any)
	if tk["title"] != "Help desk" || tk["enabled"] != true || all["files"].(map[string]any)["enabled"] != true {
		t.Fatalf("saved: %v", all)
	}
}

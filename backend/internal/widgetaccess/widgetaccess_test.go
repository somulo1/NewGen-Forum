package widgetaccess

import "testing"

func TestModeOf(t *testing.T) {
	cases := []struct{ access, see, want string }{
		{"roles", "", ByRoles},
		{"groups", "", ByGroups},
		{"", "Everyone", ByRoles},
		{"", "", ByRoles},
		{"", "Specific groups", ByGroups}, // saved before the choice existed: it was a user-group dashboard
		{"roles", "Specific groups", ByRoles},
	}
	for _, c := range cases {
		if got := ModeOf(c.access, c.see); got != c.want {
			t.Fatalf("ModeOf(%q, %q) = %q, want %q", c.access, c.see, got, c.want)
		}
	}
	for _, id := range []string{"departments", "sectors", "branches", "units"} {
		if !IsDashboard(id) {
			t.Fatalf("%s is a dashboard", id)
		}
	}
	if IsDashboard("tickets") {
		t.Fatal("tickets is not a dashboard")
	}
}

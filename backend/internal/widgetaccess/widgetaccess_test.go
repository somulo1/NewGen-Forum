package widgetaccess

import "testing"

func TestSeeRight(t *testing.T) {
	for id, want := range map[string]string{"departments": "View dashboard", "units": "View dashboard", "tickets": "View", "memos": "Read", "sessions": "Join", "notes": ""} {
		if got := SeeRight(id); got != want {
			t.Fatalf("SeeRight(%q) = %q, want %q", id, got, want)
		}
	}
}

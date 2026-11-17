package push_test

import (
	"testing"

	"hub/internal/testkit"
)

// Only sign-ins from the phone app (bound to a key kept in the phone) may receive pushes; a browser session is refused.
func TestPushNeedsAPhoneSignIn(t *testing.T) {
	e := testkit.New(t)
	owner, _ := e.NewOrg("Push Owner")
	e.MustDo(400, "POST", "/push/devices", map[string]any{"token": "fcm-token-1234567890", "platform": "windows"}, owner.Access)
	e.MustDo(403, "POST", "/push/devices", map[string]any{"token": "fcm-token-1234567890", "platform": "android"}, owner.Access)
	e.MustDo(204, "DELETE", "/push/devices", map[string]any{"token": "fcm-token-1234567890", "platform": "android"}, owner.Access)
}

package settings_test

import (
	"context"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/modules/settings"
	"hub/internal/testkit"
)

func sidOf(t *testing.T, access string) string {
	var c struct {
		jwt.RegisteredClaims
		Sid string `json:"sid"`
	}
	if _, _, err := jwt.NewParser().ParseUnverified(access, &c); err != nil {
		t.Fatal(err)
	}
	return c.Sid
}

// Inactive past the organization's limit: the next request is refused and the session ends. Activity keeps it alive.
func TestInactiveSessionsEndAtTheLimit(t *testing.T) {
	e := testkit.New(t)
	ownerU, _ := e.NewOrg("Olivia Owner")
	ctx := context.Background()
	sec := e.MustDo(200, "GET", "/settings", nil, ownerU.Access)["security"].(map[string]any)
	sec["session"] = "15"
	e.MustDo(204, "PUT", "/settings/security", sec, ownerU.Access)
	if code, _ := e.Do("PUT", "/settings/security", map[string]any{"session": "8", "retention": "12", "maxVideoMB": "50", "userQuotaMB": "5120"}, ownerU.Access); code != 400 {
		t.Fatalf("a session length not on the list: %d", code)
	}
	sid := sidOf(t, ownerU.Access)
	// clicks keep it alive
	e.MustDo(200, "POST", "/auth/active", map[string]any{}, ownerU.Access)
	e.MustDo(200, "GET", "/settings", nil, ownerU.Access)
	// 15 minutes and one second without a click or key press
	e.A.Rdb.Set(ctx, "act:"+sid, strconv.FormatInt(time.Now().Add(-15*time.Minute-time.Second).UnixMilli(), 10), time.Hour)
	code, out := e.Do("GET", "/settings", nil, ownerU.Access)
	if code != 401 || out["error"].(map[string]any)["code"] != "session_expired" {
		t.Fatalf("inactive session still works: %d %v", code, out)
	}
	if code, _ := e.Do("POST", "/auth/refresh", map[string]any{"refresh_token": ownerU.Refresh}, ""); code != 401 {
		t.Fatalf("an expired session was renewed: %d", code)
	}
}

// Today's users count each person once; a request to delete the organization waits 30 days and can be cancelled.
func TestActivityAndDeletionRequest(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	a := e.Join(ownerU, slug, "Amy", "staff")
	for i := 0; i < 3; i++ { // signing in three times is still one person today
		e.MustDo(200, "POST", "/auth/active", map[string]any{}, a.Access)
	}
	act := e.MustDo(200, "GET", "/security/activity", nil, ownerU.Access)
	if act["today"].(float64) < 2 || act["today"].(float64) > act["active"].(float64) {
		t.Fatalf("today: %v", act)
	}
	name := e.MustDo(200, "GET", "/settings", nil, ownerU.Access)["profile"].(map[string]any)["name"].(string)
	if code, _ := e.Do("POST", "/settings/deletion", map[string]any{"confirm": "wrong"}, ownerU.Access); code != 400 {
		t.Fatalf("deleted without the name: %d", code)
	}
	due := e.MustDo(200, "POST", "/settings/deletion", map[string]any{"confirm": name}, ownerU.Access)["delete_after"].(string)
	when, _ := time.Parse(time.RFC3339Nano, due)
	if d := time.Until(when); d < 29*24*time.Hour || d > 31*24*time.Hour {
		t.Fatalf("deletion date: %v", due)
	}
	if got := e.MustDo(200, "GET", "/settings", nil, ownerU.Access)["delete_after"]; got == nil {
		t.Fatal("settings do not show the deletion date")
	}
	// not due yet: nothing is erased
	if err := settings.New(e.A).EraseOrg(context.Background(), uuid.MustParse(e.OrgID(slug))); err == nil || !strings.Contains(err.Error(), "not due") {
		t.Fatalf("erased before the date: %v", err)
	}
	e.MustDo(204, "DELETE", "/settings/deletion", nil, ownerU.Access)
	if got := e.MustDo(200, "GET", "/settings", nil, ownerU.Access)["delete_after"]; got != nil {
		t.Fatal("cancel did not clear the date")
	}
}

// When the 30 days are over, every record of the organization goes (including its logs and keys), and other organizations
// are untouched.
func TestOrganizationIsErasedWhenDue(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	e.Join(ownerU, slug, "Amy", "staff")
	otherU, otherSlug := e.NewOrg("Other Owner")
	org, other := uuid.MustParse(e.OrgID(slug)), uuid.MustParse(e.OrgID(otherSlug))
	_ = otherU
	name := e.MustDo(200, "GET", "/settings", nil, ownerU.Access)["profile"].(map[string]any)["name"].(string)
	e.MustDo(200, "POST", "/settings/deletion", map[string]any{"confirm": name}, ownerU.Access)

	root, err := pgx.Connect(context.Background(), "postgres://hubadmin@/hub?host=/tmp&port=5433")
	if err != nil {
		t.Skip("superuser not reachable: ", err)
	}
	defer root.Close(context.Background())
	_, _ = root.Exec(context.Background(), "set session_replication_role = replica")
	if _, err := root.Exec(context.Background(), `update hub.orgs set delete_after = now() - interval '1 minute' where id=$1`, org); err != nil {
		t.Skip("cannot bring the date forward: ", err)
	}
	m := settings.New(e.A)
	due, err := m.DeletionDue(context.Background())
	if err != nil || !slices.Contains(due, org) || slices.Contains(due, other) {
		t.Fatalf("due: %v %v", due, err)
	}
	if err := m.EraseOrg(context.Background(), org); err != nil {
		t.Fatal(err)
	}
	count := func(id uuid.UUID, table string) int {
		var n int
		_ = root.QueryRow(context.Background(), `select count(*) from hub.`+table+` where org_id=$1`, id).Scan(&n)
		return n
	}
	for _, tbl := range []string{"people", "audit_logs", "org_keys", "conversations"} {
		if n := count(org, tbl); n != 0 {
			t.Fatalf("%s still holds %d rows of the erased organization", tbl, n)
		}
	}
	if count(other, "people") == 0 || count(other, "org_keys") == 0 {
		t.Fatal("another organization lost data")
	}
	var status string
	_ = root.QueryRow(context.Background(), `select status from hub.orgs where id=$1`, org).Scan(&status)
	if status != "deleted" {
		t.Fatalf("status %q", status)
	}
}

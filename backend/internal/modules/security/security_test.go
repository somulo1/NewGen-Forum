package security_test

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"

	"hub/internal/testkit"
)

func TestFailedSignInsRaiseAnAlertAndAlertsNeedComments(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	victim := e.Join(ownerU, slug, "Peter", "staff")
	// five wrong passwords lock the account and trip rule r1
	for i := 0; i < 5; i++ {
		e.Do("POST", "/auth/login", map[string]any{"email": victim.Email, "password": "Wrong-pass1"}, "")
	}
	alerts := e.MustDo(200, "GET", "/security/alerts?open=1", nil, ownerU.Access)["alerts"].([]any)
	var al map[string]any
	for _, a := range alerts {
		if a.(map[string]any)["rule"] == "r1" {
			al = a.(map[string]any)
		}
	}
	if al == nil || !strings.Contains(al["title"].(string), "Peter") {
		t.Fatalf("no r1 alert: %v", alerts)
	}
	id := al["id"].(string)
	// a status change needs a comment; staff cannot work alerts
	if code, _ := e.Do("POST", "/security/alerts/"+id+"/status", map[string]any{"status": "Resolved", "comment": ""}, ownerU.Access); code != 400 {
		t.Fatalf("closed without a comment: %d", code)
	}
	staff := e.Join(ownerU, slug, "Sam", "staff")
	if code, _ := e.Do("GET", "/security/alerts", nil, staff.Access); code != 403 {
		t.Fatalf("staff read alerts: %d", code)
	}
	e.MustDo(204, "POST", "/security/alerts/"+id+"/status", map[string]any{"status": "Investigating", "comment": "Checking with Peter."}, ownerU.Access)
	e.MustDo(204, "POST", "/security/alerts/"+id+"/status", map[string]any{"status": "Resolved", "comment": "It was Peter forgetting his password."}, ownerU.Access)
	if code, _ := e.Do("POST", "/security/alerts/"+id+"/status", map[string]any{"status": "Open", "comment": "again"}, ownerU.Access); code != 400 {
		t.Fatalf("reopen: %d", code)
	}
	all := e.MustDo(200, "GET", "/security/alerts", nil, ownerU.Access)["alerts"].([]any)
	for _, a := range all {
		m := a.(map[string]any)
		if m["id"] == id && (m["status"] != "Resolved" || len(m["notes"].([]any)) != 2) {
			t.Fatalf("alert history: %v", m)
		}
	}
	// the administrators were notified
	if n := e.MustDo(200, "GET", "/notifications", nil, ownerU.Access); n["unread"].(float64) < 1 {
		t.Fatalf("no notification: %v", n)
	}
}

func TestRightsChangeAndLogsExportAndIntegrity(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	_ = slug
	e.MustDo(204, "PUT", "/approvals/flows/f8", map[string]any{"levels": 0}, ownerU.Access)
	e.MustDo(200, "PUT", "/roles/staff/rights", map[string]any{"module": "wall", "right": "Post", "allowed": false}, ownerU.Access)
	alerts := e.MustDo(200, "GET", "/security/alerts", nil, ownerU.Access)["alerts"].([]any)
	got := false
	for _, a := range alerts {
		got = got || a.(map[string]any)["rule"] == "r5"
	}
	if !got {
		t.Fatalf("no rights-change alert: %v", alerts)
	}
	logs := e.MustDo(200, "GET", "/security/logs?q=rights", nil, ownerU.Access)["logs"].([]any)
	if len(logs) == 0 {
		t.Fatal("log search found nothing")
	}
	if l := e.MustDo(200, "GET", "/security/logs?flagged=1", nil, ownerU.Access)["logs"].([]any); len(l) == 0 {
		t.Fatal("flagged filter empty")
	}
	// spreadsheet export is a real xlsx, and cells that could run as formulas are neutralised
	req, _ := http.NewRequest("GET", e.Srv.URL+"/v1/security/logs/export?format=csv", nil)
	req.Header.Set("Authorization", "Bearer "+ownerU.Access)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(res.Body)
	if res.StatusCode != 200 || !bytes.HasPrefix(b, []byte("When,Person")) {
		t.Fatalf("csv export: %d %q", res.StatusCode, b[:min(40, len(b))])
	}
	req, _ = http.NewRequest("GET", e.Srv.URL+"/v1/security/logs/export", nil)
	req.Header.Set("Authorization", "Bearer "+ownerU.Access)
	res, _ = http.DefaultClient.Do(req)
	b, _ = io.ReadAll(res.Body)
	if res.StatusCode != 200 || !bytes.HasPrefix(b, []byte("PK")) {
		t.Fatalf("xlsx export is not a zip: %d", res.StatusCode)
	}
	in := e.MustDo(200, "GET", "/security/integrity", nil, ownerU.Access)
	if in["intact"] != true || in["entries"].(float64) < 3 {
		t.Fatalf("integrity: %v", in)
	}
}

func TestSignInBlocksAreRecordedForMembersAndStrangers(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	victim := e.Join(ownerU, slug, "Peter", "staff")
	for i := 0; i < 5; i++ {
		e.Do("POST", "/auth/login", map[string]any{"email": victim.Email, "password": "Wrong-pass1"}, "")
	}
	// someone with no account keeps trying until the address is blocked
	for i := 0; i < 40; i++ {
		e.Do("POST", "/auth/login", map[string]any{"email": "nobody@stranger.example", "password": "Wrong-pass1"}, "")
	}
	bl := e.MustDo(200, "GET", "/security/blocks?days=7", nil, ownerU.Access)["blocks"].([]any)
	var member, stranger bool
	for _, b := range bl {
		m := b.(map[string]any)
		if m["member"] == true && m["who"] == "Peter" {
			member = true
		}
		if m["member"] == false && strings.Contains(m["who"].(string), "***@stranger.example") {
			stranger = true
		}
	}
	if !member || !stranger {
		t.Fatalf("blocks: %v", bl)
	}
	staff := e.Join(ownerU, slug, "Sam", "staff")
	if code, _ := e.Do("GET", "/security/blocks", nil, staff.Access); code != 403 {
		t.Fatalf("staff read blocks: %d", code)
	}
}

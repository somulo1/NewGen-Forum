package settings_test

import (
	"bytes"
	"context"
	"github.com/jackc/pgx/v5"
	"testing"

	"github.com/google/uuid"

	"hub/internal/modules/settings"
	"hub/internal/testkit"
)

func TestSettingsAreEncryptedValidatedAndGuarded(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	staff := e.Join(ownerU, slug, "Sam", "staff")

	all := e.MustDo(200, "GET", "/settings", nil, ownerU.Access)
	if all["profile"].(map[string]any)["name"] == "" || len(all["plans"].([]any)) != 3 {
		t.Fatalf("settings: %v", all)
	}
	prof := map[string]any{"name": "Acme Holdings", "short": "Acme", "logo": "", "industry": "Banking", "size": "201–500", "website": "", "email": "ops@acme.test", "phone": "", "address": "1 Main St", "city": "Nairobi", "country": "Kenya",
		"timezone": "Africa/Nairobi", "language": "English", "currency": "KES", "regNo": "C.123", "taxId": "P000111222Z"}
	e.MustDo(204, "PUT", "/settings/profile", prof, ownerU.Access)
	if code, _ := e.Do("PUT", "/settings/profile", prof, staff.Access); code != 403 {
		t.Fatalf("staff edited the profile: %d", code)
	}
	bad := map[string]any{}
	for k, v := range prof {
		bad[k] = v
	}
	bad["timezone"] = "Mars/Olympus"
	if code, _ := e.Do("PUT", "/settings/profile", bad, ownerU.Access); code != 400 {
		t.Fatalf("bad timezone: %d", code)
	}
	// the tax id is not readable in the database
	var raw []byte
	if err := e.A.DB.WithOrg(context.Background(), uuid.MustParse(e.OrgID(slug)), func(tx pgxTx) error {
		return tx.QueryRow(context.Background(), `select data_enc from hub.org_settings where section='profile'`).Scan(&raw)
	}); err != nil {
		t.Fatal(err)
	}
	if string(raw) == "" || containsBytes(raw, "P000111222Z") {
		t.Fatal("settings are stored readable")
	}
	// subscription: seats must cover the people already in, and fit the plan
	sub := map[string]any{"plan": "starter", "cycle": "monthly", "seats": 600, "autoRenew": true, "renews": "", "status": "", "cancelAt": "", "addons": map[string]bool{"video": true}}
	if code, _ := e.Do("PUT", "/settings/subscription", sub, ownerU.Access); code != 400 {
		t.Fatalf("seats beyond the plan: %d", code)
	}
	sub["seats"] = 1
	if code, _ := e.Do("PUT", "/settings/subscription", sub, ownerU.Access); code != 409 {
		t.Fatalf("seats below people in use: %d", code)
	}
	sub["seats"] = 2 // owner + Sam
	e.MustDo(200, "PUT", "/settings/subscription", sub, ownerU.Access)
	// the plan is full, so no one else can be added
	if code, body := e.Do("POST", "/people", map[string]any{"people": []any{map[string]any{"first_name": "X", "last_name": "Y", "email": testkit.Uniq(), "phone": "+254700000001", "date_of_birth": "1990-04-05", "gender": "Male"}}}, ownerU.Access); code != 402 {
		t.Fatalf("seat limit not enforced: %d %v", code, body)
	}
	// billing keeps only the last four digits
	if code, _ := e.Do("PUT", "/settings/billing", map[string]any{"cardBrand": "Visa", "cardLast4": "4242424242424242", "cardExp": "08/28"}, ownerU.Access); code != 400 {
		t.Fatalf("full card number accepted: %d", code)
	}
	e.MustDo(204, "PUT", "/settings/billing", map[string]any{"cardBrand": "Visa", "cardLast4": "4242", "cardExp": "08/28", "holder": "Olivia", "contactEmail": "bill@acme.test", "vat": "0123456A"}, ownerU.Access)
	if code, _ := e.Do("GET", "/settings", nil, staff.Access); code != 403 {
		t.Fatalf("staff can open settings: %d", code)
	}
	// invoices and the tax invoice
	m := settings.New(e.A)
	if _, err := m.IssueInvoice(context.Background(), uuid.MustParse(e.OrgID(slug)), 1160, "KES", "Paid"); err != nil {
		t.Fatal(err)
	}
	inv := e.MustDo(200, "GET", "/billing/invoices", nil, ownerU.Access)["invoices"].([]any)
	if len(inv) != 1 {
		t.Fatalf("invoices: %v", inv)
	}
	etr := e.MustDo(200, "GET", "/billing/invoices/"+inv[0].(map[string]any)["no"].(string)+"/etr", nil, ownerU.Access)
	if etr["gross"].(float64) != 1160 || etr["vat"].(float64) != 160 || etr["fiscal"].(map[string]any)["integrated"] != false {
		t.Fatalf("etr: %v", etr)
	}
	if code, _ := e.Do("GET", "/billing/invoices", nil, staff.Access); code != 403 {
		t.Fatalf("staff read invoices: %d", code)
	}
}

type pgxTx = pgx.Tx

func containsBytes(b []byte, s string) bool { return bytes.Contains(b, []byte(s)) }

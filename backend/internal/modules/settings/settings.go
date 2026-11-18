// Package settings is the organization's profile, security policy, subscription, billing details and invoices.
package settings

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/authz"
	"hub/internal/httpx"
	"hub/internal/orgsettings"
	"hub/internal/realtime"
	"hub/internal/tenant"
)

type Module struct{ A *app.App }

func New(a *app.App) *Module { return &Module{A: a} }

type Profile struct {
	Name     string `json:"name"`
	Short    string `json:"short"`
	Logo     string `json:"logo"`
	Industry string `json:"industry"`
	Size     string `json:"size"`
	Website  string `json:"website"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`
	City     string `json:"city"`
	Country  string `json:"country"`
	Timezone string `json:"timezone"`
	Language string `json:"language"`
	Currency string `json:"currency"`
	RegNo    string `json:"regNo"`
	TaxID    string `json:"taxId"`
}

type Billing struct {
	CardBrand    string `json:"cardBrand"`
	CardLast4    string `json:"cardLast4"`
	CardExp      string `json:"cardExp"`
	Holder       string `json:"holder"`
	ContactEmail string `json:"contactEmail"`
	VAT          string `json:"vat"`
}

type All struct {
	Profile      Profile                  `json:"profile"`
	Security     orgsettings.Security     `json:"security"`
	Subscription orgsettings.Subscription `json:"subscription"`
	Billing      Billing                  `json:"billing"`
	Plans        []orgsettings.Plan       `json:"plans"`
	Addons       []orgsettings.Addon      `json:"addons"`
	SeatsInUse   int                      `json:"seats_in_use"`
	DeleteAfter  *time.Time               `json:"delete_after,omitempty"` // when a requested deletion happens
}

var (
	emailRe = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
	cardRe  = regexp.MustCompile(`^\d{4}$`)
	expRe   = regexp.MustCompile(`^(0[1-9]|1[0-2])/\d{2}$`)
)

func (m *Module) Get(ctx context.Context, p *tenant.Principal) (*All, error) {
	out := &All{Security: orgsettings.DefaultSecurity(), Subscription: orgsettings.DefaultSubscription(), Plans: orgsettings.Plans, Addons: orgsettings.Addons}
	err := m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "settings", "View settings"); err != nil {
			return err
		}
		var name string
		if err := tx.QueryRow(ctx, `select name, delete_after from hub.orgs where id=$1`, p.Org).Scan(&name, &out.DeleteAfter); err != nil {
			return err
		}
		out.Profile = Profile{Name: name, Timezone: "Africa/Nairobi", Language: "English", Currency: "KES"}
		for sec, dst := range map[string]any{"profile": &out.Profile, "security": &out.Security, "subscription": &out.Subscription, "billing": &out.Billing} {
			if err := orgsettings.Get(ctx, m.A, tx, p.Org, sec, dst); err != nil {
				return err
			}
		}
		// billing is shown to people who manage the subscription only
		if ok, _ := authz.Can(ctx, tx, p.Person, "settings", "Manage subscription"); !ok {
			out.Billing = Billing{}
		}
		return tx.QueryRow(ctx, `select count(*) from hub.people where status <> 'deleted'`).Scan(&out.SeatsInUse)
	})
	return out, err
}

func (m *Module) SetProfile(ctx context.Context, p *tenant.Principal, in Profile) error {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" || len(in.Name) > 120 || len(in.Short) > 20 {
		return httpx.Bad("Enter the organization's name (up to 120 characters) and a short name (up to 20).")
	}
	if in.Email != "" && !emailRe.MatchString(in.Email) {
		return httpx.Bad("Enter a valid contact email.")
	}
	if _, err := time.LoadLocation(in.Timezone); err != nil || in.Timezone == "" {
		return httpx.Bad("Choose a valid time zone.")
	}
	if len(in.Currency) != 3 {
		return httpx.Bad("Currency is a three letter code like KES.")
	}
	if in.Logo != "" {
		// a data URL of a PNG, JPEG, WebP or SVG image of at most 1 MB
		head, data, ok := strings.Cut(in.Logo, ";base64,")
		if !ok || !slices.Contains([]string{"data:image/png", "data:image/jpeg", "data:image/webp", "data:image/svg+xml"}, head) {
			return httpx.Bad("The logo must be a PNG, JPEG, WebP or SVG image.")
		}
		if base64.StdEncoding.DecodedLen(len(data)) > 1<<20+2 {
			return httpx.Bad("The logo can be up to 1 MB.")
		}
		if _, err := base64.StdEncoding.DecodeString(data); err != nil {
			return httpx.Bad("The logo file could not be read.")
		}
	}
	for _, s := range []string{in.Industry, in.Size, in.Website, in.Phone, in.Address, in.City, in.Country, in.Language, in.RegNo, in.TaxID} {
		if len(s) > 200 {
			return httpx.Bad("One of the fields is too long.")
		}
	}
	return m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "settings", "Edit organization profile"); err != nil {
			return err
		}
		if err := orgsettings.Set(ctx, m.A, tx, p.Org, "profile", in); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `update hub.orgs set name=$2, updated_at=now(), ver=ver+1, mac=hub.pending_mac() where id=$1`, p.Org, in.Name); err != nil {
			return err
		}
		if err := m.A.Sign.Seal(ctx, tx, "orgs", "id=$1", p.Org); err != nil {
			return err
		}
		return m.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Updated the organization profile", "Settings", p.IP, map[string]any{"name": in.Name})
	})
}

func (m *Module) SetSecurity(ctx context.Context, p *tenant.Principal, in orgsettings.Security) error {
	if !slices.Contains(orgsettings.SessionChoices, in.Session) {
		return httpx.Bad("Choose when inactive people are signed out from the list.")
	}
	if r := in.Retention; r != "3" && r != "6" && r != "12" && r != "24" && r != "60" {
		return httpx.Bad("Choose how long to keep records from the list.")
	}
	if v := in.MaxVideoMB; v != "5" && v != "10" && v != "25" && v != "50" && v != "100" {
		return httpx.Bad("Choose a video size limit from the list.")
	}
	if in.UserQuotaMB == "" {
		in.UserQuotaMB = orgsettings.DefaultSecurity().UserQuotaMB // a client that does not know the field yet
	}
	if !slices.Contains(orgsettings.QuotaChoices, in.UserQuotaMB) {
		return httpx.Bad("Choose a storage limit from the list.")
	}
	if v := in.MaxFileMB; v != "2" && v != "5" && v != "10" && v != "25" && v != "50" {
		return httpx.Bad("Choose a file size limit from the list.")
	}
	if len(in.FileTypes) > 300 {
		return httpx.Bad("Too many file types.")
	}
	if len(in.Domains) > 500 {
		return httpx.Bad("Too many domains.")
	}
	err := m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "settings", "Edit organization profile"); err != nil {
			return err
		}
		if err := orgsettings.Set(ctx, m.A, tx, p.Org, "security", in); err != nil {
			return err
		}
		return m.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Changed the organization's security settings", "Settings", p.IP, in)
	})
	if err == nil {
		// every server applies the new inactivity limit on the next request
		m.A.Rdb.Set(ctx, "idle:"+p.Org.String(), strconv.Itoa(orgsettings.SessionMinutes(in.Session)), 10*time.Minute)
		orgsettings.ForgetIdle(p.Org)
		m.A.Publish(ctx, realtime.Event{Type: "widgets.changed", Channel: realtime.OrgChannel(p.Org), Org: p.Org.String(), Actor: p.Person.String()})
	}
	return err
}

// Widgets is the organization-wide configuration of the home widgets (titles, who sees them, who manages them, and whether the
// organization has switched each one on). A widget the organization switched on is what its people get: someone who does not manage
// widget settings is sent only the switched-on ones, so a widget that was never switched on does not even reach their screen.
func (m *Module) Widgets(ctx context.Context, p *tenant.Principal) (map[string]any, error) {
	out := map[string]any{}
	err := m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := orgsettings.Get(ctx, m.A, tx, p.Org, "widgets", &out); err != nil {
			return err
		}
		manage, err := authz.Can(ctx, tx, p.Person, "home", "Manage widget settings")
		if err != nil {
			return err
		}
		if !manage {
			for id, v := range out {
				if c, ok := v.(map[string]any); !ok || c["enabled"] != true {
					delete(out, id) // only a widget the organization explicitly switched on reaches its people
				}
			}
		}
		return nil
	})
	return out, err
}

var widgetID = regexp.MustCompile(`^[a-z][a-z0-9_]{0,31}$`)

// validWidgets checks the shape of the configuration before it is saved: a widget id maps to an object, `enabled` is a yes or no, and
// who may see it is one of the known choices.
func validWidgets(in map[string]any) error {
	if len(in) > 80 {
		return httpx.Bad("That is too many widgets.")
	}
	for id, v := range in {
		c, ok := v.(map[string]any)
		if !widgetID.MatchString(id) || !ok {
			return httpx.Bad("A widget setting is not valid.")
		}
		if e, has := c["enabled"]; has {
			if _, isBool := e.(bool); !isBool {
				return httpx.Bad("Switched on must be yes or no.")
			}
		}
		if see, has := c["see"]; has && see != "Everyone" && see != "Specific groups" {
			return httpx.Bad("Choose who can see the widget from the list.")
		}
		if acc, has := c["access"]; has && acc != "roles" && acc != "groups" {
			return httpx.Bad("Choose whether roles or user groups open the dashboard.")
		}
	}
	return nil
}

func (m *Module) SetWidgets(ctx context.Context, p *tenant.Principal, in map[string]any) error {
	if raw, _ := json.Marshal(in); len(raw) > 32<<10 {
		return httpx.Bad("That is too much widget configuration.")
	}
	if err := validWidgets(in); err != nil {
		return err
	}
	var before map[string]any
	err := m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "home", "Manage widget settings"); err != nil {
			return err
		}
		_ = orgsettings.Get(ctx, m.A, tx, p.Org, "widgets", &before)
		if err := orgsettings.Set(ctx, m.A, tx, p.Org, "widgets", in); err != nil {
			return err
		}
		// the audit log says which widgets were switched on or off, not only that something changed
		var on, off []string
		for id, v := range in {
			now := v.(map[string]any)["enabled"] == true
			was := false
			if b, ok := before[id].(map[string]any); ok {
				was = b["enabled"] == true
			}
			if now && !was {
				on = append(on, id)
			} else if !now && was {
				off = append(off, id)
			}
		}
		for id, v := range before {
			if _, still := in[id]; !still {
				if b, ok := v.(map[string]any); ok && b["enabled"] == true {
					off = append(off, id)
				}
			}
		}
		slices.Sort(on)
		slices.Sort(off)
		return m.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Changed the home widget settings", "Settings", p.IP, map[string]any{"switched_on": on, "switched_off": off})
	})
	if err != nil {
		return err
	}
	// everyone signed in to this organization is told, so the change shows on their screen now, not at their next sign-in
	m.A.Publish(ctx, realtime.Event{Type: "widgets.changed", Channel: realtime.OrgChannel(p.Org), Org: p.Org.String(), Actor: p.Person.String()})
	return nil
}

// monthly price of a subscription in major units
func price(s orgsettings.Subscription) int {
	p, _ := orgsettings.FindPlan(s.Plan)
	total := p.Price * s.Seats
	for _, a := range orgsettings.Addons {
		if s.Addons[a.ID] {
			total += a.Price * 10
		}
	}
	return total
}

func (m *Module) SetSubscription(ctx context.Context, p *tenant.Principal, in orgsettings.Subscription) (*All, error) {
	plan, ok := orgsettings.FindPlan(in.Plan)
	if !ok {
		return nil, httpx.Bad("Choose one of the plans.")
	}
	if in.Cycle != "monthly" && in.Cycle != "annual" {
		return nil, httpx.Bad("Choose monthly or annual billing.")
	}
	if in.Seats < 1 || in.Seats > plan.Seats {
		return nil, httpx.Bad(fmt.Sprintf("The %s plan covers 1 to %d seats.", plan.Name, plan.Seats))
	}
	for k := range in.Addons {
		found := false
		for _, a := range orgsettings.Addons {
			found = found || a.ID == k
		}
		if !found {
			return nil, httpx.Bad("Unknown add-on.")
		}
	}
	err := m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "settings", "Manage subscription"); err != nil {
			return err
		}
		var used int
		if err := tx.QueryRow(ctx, `select count(*) from hub.people where status <> 'deleted'`).Scan(&used); err != nil {
			return err
		}
		if in.Seats < used {
			return httpx.Err(409, "seats_in_use", fmt.Sprintf("%d people are in the organization. Remove some before reducing seats below that.", used))
		}
		cur := orgsettings.DefaultSubscription()
		if err := orgsettings.Get(ctx, m.A, tx, p.Org, "subscription", &cur); err != nil {
			return err
		}
		in.Status, in.Renews = cur.Status, cur.Renews
		if in.Renews == "" {
			in.Renews = time.Now().AddDate(0, 1, 0).Format("2006-01-02")
		}
		if in.CancelAt != "" {
			if _, err := time.Parse("2006-01-02", in.CancelAt); err != nil {
				return httpx.Bad("Cancel date looks like 2026-12-31.")
			}
			in.Status = "Cancelling"
		} else if cur.Status == "Cancelling" {
			in.Status = "Active"
		}
		if err := orgsettings.Set(ctx, m.A, tx, p.Org, "subscription", in); err != nil {
			return err
		}
		return m.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, fmt.Sprintf("Changed the subscription to %s, %d seats", plan.Name, in.Seats), "Settings", p.IP, in)
	})
	if err != nil {
		return nil, err
	}
	return m.Get(ctx, p)
}

func (m *Module) SetBilling(ctx context.Context, p *tenant.Principal, in Billing) error {
	if in.CardLast4 != "" && !cardRe.MatchString(in.CardLast4) {
		return httpx.Bad("Only the last four digits are kept.")
	}
	if in.CardExp != "" && !expRe.MatchString(in.CardExp) {
		return httpx.Bad("The expiry looks like 08/28.")
	}
	if in.ContactEmail != "" && !emailRe.MatchString(in.ContactEmail) {
		return httpx.Bad("Enter a valid billing email.")
	}
	if len(in.Holder) > 100 || len(in.VAT) > 40 || len(in.CardBrand) > 20 {
		return httpx.Bad("One of the fields is too long.")
	}
	return m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "settings", "Manage subscription"); err != nil {
			return err
		}
		if err := orgsettings.Set(ctx, m.A, tx, p.Org, "billing", in); err != nil {
			return err
		}
		return m.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Updated billing details", "Settings", p.IP, nil)
	})
}

// ---- invoices ----

type Invoice struct {
	No     string `json:"no"`
	Date   string `json:"date"`
	Amount int64  `json:"amount"` // major units
	Status string `json:"status"`
}

func (m *Module) Invoices(ctx context.Context, p *tenant.Principal) ([]Invoice, error) {
	out := []Invoice{}
	err := m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "settings", "Manage subscription"); err != nil {
			return err
		}
		rs, err := tx.Query(ctx, `select no, issued_on::text, amount_minor, status from hub.invoices order by issued_on desc, no desc limit 120`)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var i Invoice
			var minor int64
			if err := rs.Scan(&i.No, &i.Date, &minor, &i.Status); err != nil {
				return err
			}
			i.Amount = minor / 100
			out = append(out, i)
		}
		return rs.Err()
	})
	return out, err
}

const vat = 0.16

type Line struct {
	Desc  string  `json:"desc"`
	Qty   int     `json:"qty"`
	Net   float64 `json:"net"`
	VAT   float64 `json:"vat"`
	Gross float64 `json:"gross"`
}

type ETR struct {
	Kind      string  `json:"kind"`
	Title     string  `json:"title"`
	InvoiceNo string  `json:"invoice_no"`
	Date      string  `json:"date"`
	Status    string  `json:"status"`
	Seller    Seller  `json:"seller"`
	Buyer     Profile `json:"buyer"`
	BuyerVAT  string  `json:"buyer_vat"`
	Lines     []Line  `json:"lines"`
	Net       float64 `json:"net"`
	VAT       float64 `json:"vat"`
	Gross     float64 `json:"gross"`
	Currency  string  `json:"currency"`
	Fiscal    Fiscal  `json:"fiscal"`
}

type Seller struct {
	Name, PIN, VAT, Address, Phone, Email string
}

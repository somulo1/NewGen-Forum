// Package settings is the organization's profile, security policy, subscription, billing details and invoices.
package settings

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/authz"
	"hub/internal/crypto"
	"hub/internal/httpx"
	"hub/internal/orgsettings"
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
		if err := tx.QueryRow(ctx, `select name from hub.orgs where id=$1`, p.Org).Scan(&name); err != nil {
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
	if len(in.Logo) > 700<<10 || (in.Logo != "" && !strings.HasPrefix(in.Logo, "data:image/")) {
		return httpx.Bad("The logo must be an image under 500 KB.")
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
	if h := in.Session; h != "1" && h != "4" && h != "8" && h != "12" && h != "24" && h != "168" {
		return httpx.Bad("Choose a session length from the list.")
	}
	if r := in.Retention; r != "3" && r != "6" && r != "12" && r != "24" && r != "60" {
		return httpx.Bad("Choose how long to keep records from the list.")
	}
	if len(in.Domains) > 500 {
		return httpx.Bad("Too many domains.")
	}
	return m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "settings", "Edit organization profile"); err != nil {
			return err
		}
		if err := orgsettings.Set(ctx, m.A, tx, p.Org, "security", in); err != nil {
			return err
		}
		return m.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Changed the organization's security settings", "Settings", p.IP, in)
	})
}

// Widgets is the organization-wide configuration of the home widgets (titles, who sees them, who manages them).
func (m *Module) Widgets(ctx context.Context, p *tenant.Principal) (map[string]any, error) {
	out := map[string]any{}
	err := m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error { return orgsettings.Get(ctx, m.A, tx, p.Org, "widgets", &out) })
	return out, err
}

func (m *Module) SetWidgets(ctx context.Context, p *tenant.Principal, in map[string]any) error {
	if raw, _ := json.Marshal(in); len(raw) > 32<<10 {
		return httpx.Bad("That is too much widget configuration.")
	}
	return m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "home", "Manage widget settings"); err != nil {
			return err
		}
		if err := orgsettings.Set(ctx, m.A, tx, p.Org, "widgets", in); err != nil {
			return err
		}
		return m.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Changed the home widget settings", "Settings", p.IP, nil)
	})
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

type Fiscal struct {
	CUSerial   string `json:"cu_serial"`
	CUInvoice  string `json:"cu_invoice"`
	Signature  string `json:"signature"`
	Integrated bool   `json:"integrated"` // true only when a fiscal device / KRA eTIMS issued these numbers
	Note       string `json:"note"`
}

func split(gross float64) (net, v float64) {
	net = math.Round(gross/(1+vat)*100) / 100
	return net, math.Round((gross-net)*100) / 100
}

// ETRFor builds the tax invoice or receipt for one invoice. The fiscal numbers are derived with a keyed hash so they are
// stable and cannot be guessed, but they are NOT legal fiscal numbers until a certified fiscal device or KRA eTIMS issues them.
func (m *Module) ETRFor(ctx context.Context, p *tenant.Principal, no string) (*ETR, error) {
	var out *ETR
	err := m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "settings", "Manage subscription"); err != nil {
			return err
		}
		var inv Invoice
		var minor int64
		if err := tx.QueryRow(ctx, `select no, issued_on::text, amount_minor, status from hub.invoices where no=$1`, no).Scan(&inv.No, &inv.Date, &minor, &inv.Status); err != nil {
			return httpx.ErrNotFound
		}
		sub := orgsettings.DefaultSubscription()
		var prof Profile
		var bill Billing
		_ = tx.QueryRow(ctx, `select name from hub.orgs where id=$1`, p.Org).Scan(&prof.Name)
		for sec, dst := range map[string]any{"subscription": &sub, "profile": &prof, "billing": &bill} {
			if err := orgsettings.Get(ctx, m.A, tx, p.Org, sec, dst); err != nil {
				return err
			}
		}
		plan, _ := orgsettings.FindPlan(sub.Plan)
		gross := float64(minor) / 100
		addons := 0.0
		var lines []Line
		for _, a := range orgsettings.Addons {
			if sub.Addons[a.ID] {
				g := float64(a.Price * 10)
				addons += g
				n, v := split(g)
				lines = append(lines, Line{a.Name + " (monthly)", 1, n, v, g})
			}
		}
		pg := math.Max(0, gross-addons)
		n, v := split(pg)
		cycle := "monthly"
		if sub.Cycle == "annual" {
			cycle = "yearly"
		}
		lines = append([]Line{{fmt.Sprintf("Hub %s plan · %d seats (%s)", plan.Name, sub.Seats, cycle), 1, n, v, pg}}, lines...)
		e := &ETR{Kind: "invoice", Title: "TAX INVOICE", InvoiceNo: inv.No, Date: inv.Date, Status: inv.Status, Buyer: prof, BuyerVAT: bill.VAT, Lines: lines, Currency: prof.Currency,
			Seller: Seller{"Hub Platforms Limited", "P052198341K", "0123456A", "Westlands Business Park, Nairobi", "+254 20 000 0000", "billing@hub.example"}}
		if inv.Status == "Paid" {
			e.Kind, e.Title = "receipt", "PAYMENT RECEIPT"
		}
		for _, l := range lines {
			e.Net += l.Net
			e.VAT += l.VAT
			e.Gross += l.Gross
		}
		h := crypto.MAC(m.A.Cfg.MACKey, []byte("hub/etr/v1"), p.Org[:], []byte(no))
		e.Fiscal = Fiscal{CUSerial: fmt.Sprintf("KRAMW%010d", beUint(h[0:4])%10000000000), CUInvoice: fmt.Sprintf("00%06d/%08d", beUint(h[4:8])%900000+100000, beUint(h[8:12])%90000+10000),
			Signature: fmt.Sprintf("%X-%X-%X-%X", h[12:16], h[16:20], h[20:24], h[24:28]), Integrated: false,
			Note: "Layout numbers only. Connect a certified fiscal device or KRA eTIMS to issue legal fiscal numbers."}
		out = e
		return nil
	})
	return out, err
}

func beUint(b []byte) uint64 {
	return uint64(b[0])<<24 | uint64(b[1])<<16 | uint64(b[2])<<8 | uint64(b[3])
}

// IssueInvoice records a charge. Billing integrations (the card gateway, the renewal job) call it; people do not.
func (m *Module) IssueInvoice(ctx context.Context, org uuid.UUID, amount int64, currency, status string) (string, error) {
	return m.IssueInvoiceOn(ctx, org, amount, currency, status, "")
}

// IssueInvoiceOn is IssueInvoice with an issue date (history loading); an empty date means today.
func (m *Module) IssueInvoiceOn(ctx context.Context, org uuid.UUID, amount int64, currency, status, date string) (string, error) {
	var no string
	err := m.A.DB.WithOrg(ctx, org, func(tx pgx.Tx) error {
		var n int
		if err := tx.QueryRow(ctx, `insert into hub.counters (org_id, name, value) values ($1,'invoice',1) on conflict (org_id, name) do update set value = hub.counters.value + 1 returning value`, org).Scan(&n); err != nil {
			return err
		}
		year := time.Now().Year()
		if len(date) >= 4 {
			year, _ = strconv.Atoi(date[:4])
		}
		no = fmt.Sprintf("INV-%d-%03d", year, n)
		if _, err := tx.Exec(ctx, `insert into hub.invoices (org_id, no, issued_on, amount_minor, currency, status) values ($1,$2,coalesce(nullif($6,'')::date, current_date),$3,$4,$5)`, org, no, amount*100, currency, status, date); err != nil {
			return err
		}
		return m.A.Sign.Seal(ctx, tx, "invoices", "no=$1", no)
	})
	return no, err
}

func (m *Module) Public(chi.Router) {}

func (m *Module) Private(r chi.Router) {
	org := func(h func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error) http.HandlerFunc {
		return httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
			p, err := tenant.MustOrg(r.Context())
			if err != nil {
				return err
			}
			return h(w, r, p)
		})
	}
	r.Get("/settings", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		a, err := m.Get(r.Context(), p)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, a)
		return nil
	}))
	r.Get("/settings/widgets", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		a, err := m.Widgets(r.Context(), p)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, a)
		return nil
	}))
	r.Put("/settings/widgets", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		var in map[string]any
		if err := httpx.Decode(r, &in, 40<<10); err != nil {
			return err
		}
		if err := m.SetWidgets(r.Context(), p, in); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Put("/settings/profile", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		var in Profile
		if err := httpx.Decode(r, &in, 800<<10); err != nil {
			return err
		}
		if err := m.SetProfile(r.Context(), p, in); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Put("/settings/security", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		var in orgsettings.Security
		if err := httpx.Decode(r, &in, 8<<10); err != nil {
			return err
		}
		if err := m.SetSecurity(r.Context(), p, in); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Put("/settings/subscription", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		var in orgsettings.Subscription
		if err := httpx.Decode(r, &in, 8<<10); err != nil {
			return err
		}
		a, err := m.SetSubscription(r.Context(), p, in)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, a)
		return nil
	}))
	r.Put("/settings/billing", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		var in Billing
		if err := httpx.Decode(r, &in, 8<<10); err != nil {
			return err
		}
		if err := m.SetBilling(r.Context(), p, in); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Get("/billing/invoices", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		l, err := m.Invoices(r.Context(), p)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"invoices": l})
		return nil
	}))
	r.Get("/billing/invoices/{no}/etr", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		e, err := m.ETRFor(r.Context(), p, chi.URLParam(r, "no"))
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, e)
		return nil
	}))
}

package settings

import (
	"context"
	"fmt"
	"hub/internal/widgetaccess"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/authz"
	"hub/internal/crypto"
	"hub/internal/httpx"
	"hub/internal/orgsettings"
	"hub/internal/tenant"
)

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
	r.Get("/settings/blocks", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		var out orgsettings.Blocks
		err := m.A.DB.WithOrg(r.Context(), p.Org, func(tx pgx.Tx) error {
			var e error
			out, e = orgsettings.BlocksOf(r.Context(), m.A, tx, p.Org)
			return e
		})
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, out)
		return nil
	}))
	// the widgets (and so the dashboards) this person may use: switched on, and their role holds the widget's right
	r.Get("/me/widgets", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		var out map[string]bool
		err := m.A.DB.WithOrg(r.Context(), p.Org, func(tx pgx.Tx) error {
			var e error
			out, e = widgetaccess.Allowed(r.Context(), m.A, tx, p)
			return e
		})
		if err != nil {
			return err
		}
		ids := []string{}
		for id := range out {
			ids = append(ids, id)
		}
		httpx.JSON(w, 200, map[string]any{"allowed": ids})
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
		if err := httpx.Decode(r, &in, 1500<<10); err != nil {
			return err
		}
		if err := m.SetProfile(r.Context(), p, in); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	// {"confirm": "<organization name>"}: everything is erased 30 days later unless cancelled
	r.Post("/settings/deletion", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		var in struct {
			Confirm string `json:"confirm"`
		}
		if err := httpx.Decode(r, &in, 1<<10); err != nil {
			return err
		}
		due, err := m.RequestDeletion(r.Context(), p, in.Confirm)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"delete_after": due})
		return nil
	}))
	r.Delete("/settings/deletion", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		if err := m.CancelDeletion(r.Context(), p); err != nil {
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

// Package orgsettings reads and writes an organization's settings sections, encrypted at rest.
package orgsettings

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/crypto"
	"hub/internal/integrity"
)

type Security struct {
	MFA             bool   `json:"mfa"`
	Session         string `json:"session"` // hours
	Domains         string `json:"domains"`
	Retention       string `json:"retention"` // months
	ExternalSharing bool   `json:"externalSharing"`
}

func DefaultSecurity() Security {
	return Security{Session: "8", Retention: "12", ExternalSharing: true}
}

func cx(org uuid.UUID, section string) crypto.Context {
	return crypto.Context{Org: org.String(), Table: "org_settings", Column: "data_enc", RowID: org.String() + "/" + section}
}

// Get decodes a section into dst; when the organization never saved it, dst is left as the caller set it.
func Get(ctx context.Context, a *app.App, tx pgx.Tx, org uuid.UUID, section string, dst any) error {
	var enc []byte
	err := tx.QueryRow(ctx, `select data_enc from hub.org_settings where section=$1`, section).Scan(&enc)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	dek, _, err := a.OrgDEK(ctx, tx, org)
	if err != nil {
		return err
	}
	raw, err := crypto.FieldOpen(dek, cx(org, section), enc)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, dst)
}

func Set(ctx context.Context, a *app.App, tx pgx.Tx, org uuid.UUID, section string, v any) error {
	dek, _, err := a.OrgDEK(ctx, tx, org)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}
	enc, err := crypto.Field(dek, cx(org, section), raw)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `insert into hub.org_settings (org_id, section, data_enc) values ($1,$2,$3) on conflict (org_id, section) do update set data_enc=excluded.data_enc, updated_at=now(), `+integrity.BumpOn("org_settings"), org, section, enc); err != nil {
		return err
	}
	return a.Sign.Seal(ctx, tx, "org_settings", "section=$1", section)
}

type Subscription struct {
	Plan      string          `json:"plan"`
	Cycle     string          `json:"cycle"`
	Seats     int             `json:"seats"`
	AutoRenew bool            `json:"autoRenew"`
	Renews    string          `json:"renews"`
	Status    string          `json:"status"`
	CancelAt  string          `json:"cancelAt"`
	Addons    map[string]bool `json:"addons"`
}

func DefaultSubscription() Subscription {
	return Subscription{Plan: "business", Cycle: "monthly", Seats: 300, AutoRenew: true, Status: "Active", Addons: map[string]bool{"storage": false, "video": false, "sms": false}}
}

type Plan struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Price   int    `json:"price"` // per seat per month, in major units of the billing currency
	Seats   int    `json:"seats"`
	Storage int    `json:"storage_gb"`
}

var Plans = []Plan{{"starter", "Starter", 4, 50, 50}, {"business", "Business", 9, 500, 500}, {"enterprise", "Enterprise", 15, 5000, 5000}}

type Addon struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}

var Addons = []Addon{{"storage", "Extra storage", 20}, {"video", "Video minutes", 15}, {"sms", "SMS alerts", 10}}

func FindPlan(id string) (Plan, bool) {
	for _, p := range Plans {
		if p.ID == id {
			return p, true
		}
	}
	return Plan{}, false
}

// Seats says how many seats the organization has paid for.
func SeatLimit(ctx context.Context, a *app.App, tx pgx.Tx, org uuid.UUID) (int, error) {
	s := DefaultSubscription()
	if err := Get(ctx, a, tx, org, "subscription", &s); err != nil {
		return 0, err
	}
	p, ok := FindPlan(s.Plan)
	if !ok {
		return s.Seats, nil
	}
	return min(s.Seats, p.Seats), nil
}

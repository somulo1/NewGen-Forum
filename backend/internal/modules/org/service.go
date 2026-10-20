// Package org creates organizations and manages their structure (sectors, branches, units, departments).
package org

import (
	"context"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/access"
	"hub/internal/app"
	"hub/internal/db"
	"hub/internal/httpx"
	"hub/internal/modules/auth"
	"hub/internal/people"
)

type Service struct {
	A    *app.App
	Auth *auth.Service
}

func New(a *app.App, au *auth.Service) *Service { return &Service{A: a, Auth: au} }

var slugRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,40}$`)

type Created struct {
	ID     uuid.UUID `json:"id"`
	Slug   string    `json:"slug"`
	Name   string    `json:"name"`
	Person uuid.UUID `json:"person"`
}

// Create makes an organization with everything it needs to be usable on day one: its data key, the default roles and
// their rights, the approval flows, and the creator as its super admin. All of it is one transaction.
func (s *Service) Create(ctx context.Context, account uuid.UUID, displayName, name, slug, ip string) (*Created, error) {
	name, slug = strings.TrimSpace(name), strings.ToLower(strings.TrimSpace(slug))
	if name == "" || len(name) > 120 {
		return nil, httpx.Bad("Enter the organization's name.")
	}
	if !slugRe.MatchString(slug) {
		return nil, httpx.Bad("The web name can use lowercase letters, numbers and dashes (2 to 41 characters).")
	}
	cat := access.Default()
	org, person := app.NewID(), app.NewID()
	err := s.A.DB.Tx(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `insert into hub.orgs (id, slug, name) values ($1,$2,$3)`, org, slug, name); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "orgs", "id=$1", org); err != nil {
			return err
		}
		dek, _, err := s.A.OrgDEK(ctx, tx, org)
		if err != nil {
			return err
		}
		if err := db.SetOrg(ctx, tx, org); err != nil {
			return err
		}
		// the catalog of pages and rights, and the one role every organization starts with: the owner. Everything else
		// (roles and what they may do) is made by the organization on the Roles page.
		if err := access.Install(ctx, tx, org, func(table, where string, args ...any) error { return s.A.Sign.Seal(ctx, tx, table, where, args...) }); err != nil {
			return err
		}
		ownerRole := app.NewID()
		if _, err := tx.Exec(ctx, `insert into hub.roles (id, org_id, key, name, description, scope, locked) values ($1,$2,'owner','Owner','Full control of Hub. This role cannot be limited.','Whole organization',true)`, ownerRole, org); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "roles", "org_id=$1", org); err != nil {
			return err
		}
		// the owner role is linked to every right in the catalog from the start
		if err := access.LinkLockedRoles(ctx, tx, org, func(table, where string, args ...any) error { return s.A.Sign.Seal(ctx, tx, table, where, args...) }); err != nil {
			return err
		}
		for _, f := range cat.Flows {
			if _, err := tx.Exec(ctx, `insert into hub.approval_flows (org_id, key, module, action, levels) values ($1,$2,$3,$4,$5)`, org, f.ID, f.Mod, f.Action, f.Levels); err != nil {
				return err
			}
		}
		if err := s.A.Sign.Seal(ctx, tx, "approval_flows", "org_id=$1", org); err != nil {
			return err
		}
		// the default ticket categories every organization starts with (each can be changed or replaced per unit)
		for _, c := range []struct {
			n string
			h int
			p string
		}{{"Network & Wi-Fi", 4, "Medium"}, {"Hardware", 48, "Medium"}, {"Software bug", 72, "Medium"}, {"Access request", 8, "Medium"}, {"HR request", 120, "Low"}} {
			if _, err := tx.Exec(ctx, `insert into hub.ticket_categories (org_id, unit_type, name, tat_hours, priority) values ($1,'desk',$2,$3,$4)`, org, c.n, c.h, c.p); err != nil {
				return err
			}
		}
		if err := s.A.Sign.Seal(ctx, tx, "ticket_categories", "org_id=$1", org); err != nil {
			return err
		}
		// the creator becomes the first person and super admin
		email, err := s.Auth.AccountEmail(ctx, tx, account)
		if err != nil {
			return err
		}
		first, last := splitName(displayName)
		emailEnc, err := people.Enc(dek, org, person, people.ColEmail, email)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `insert into hub.people (id, org_id, account_id, first_name, last_name, title, email_idx, email_enc, joined_on) values ($1,$2,$3,$4,$5,'Owner',$6,$7,current_date)`,
			person, org, account, first, last, people.EmailIdx(s.A.Cfg.BlindKey, org, email), emailEnc); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "people", "id=$1", person); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `insert into hub.person_roles (org_id, person_id, role_id) values ($1,$2,$3)`, org, person, ownerRole); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "person_roles", "org_id=$1", org); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `insert into hub.account_orgs (account_id, org_id) values ($1,$2)`, account, org); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "account_orgs", "account_id=$1 and org_id=$2", account, org); err != nil {
			return err
		}
		return s.A.Audit(ctx, tx, org, &person, displayName, "Created the organization", "Settings", ip, map[string]any{"org": org, "slug": slug})
	})
	if err != nil {
		if strings.Contains(err.Error(), "orgs_slug_key") {
			return nil, httpx.Err(409, "slug_taken", "That web name is already taken. Choose another.")
		}
		return nil, err
	}
	return &Created{ID: org, Slug: slug, Name: name, Person: person}, nil
}

func sendBatch(ctx context.Context, tx pgx.Tx, b *pgx.Batch) error {
	if b.Len() == 0 {
		return nil
	}
	br := tx.SendBatch(ctx, b)
	for i := 0; i < b.Len(); i++ {
		if _, err := br.Exec(); err != nil {
			_ = br.Close()
			return err
		}
	}
	return br.Close()
}

func splitName(n string) (string, string) {
	f := strings.Fields(n)
	switch len(f) {
	case 0:
		return "Member", ""
	case 1:
		return f[0], ""
	}
	return f[0], strings.Join(f[1:], " ")
}

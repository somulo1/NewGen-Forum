package main

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/config"
	"hub/internal/crypto"
	"hub/internal/e2ee"
	"hub/internal/mail"
	"hub/internal/modules/approvals"
	"hub/internal/modules/auth"
	"hub/internal/modules/keys"
	"hub/internal/modules/people"
	"hub/internal/tenant"
)

// member adds a ready-to-use person to an organization, for development and testing: an active account with the password you
// give (the sign-up strength rules are skipped), a person in the organization (with the role if one is named), and a device
// whose keys are derived from the email, exactly as the browser derives them in development, so the person can be chatted with
// and signed in as at once, with no invitation step. It refuses to run outside development.
func member(ctx context.Context, a *app.App, cfg *config.Config, slug, name, email, password, roleKey string) error {
	if cfg.Env != "dev" {
		return errors.New("member only runs in development (HUB_ENV=dev): it sets passwords directly")
	}
	email = strings.ToLower(strings.TrimSpace(email))
	name = strings.TrimSpace(name)
	if email == "" || password == "" || name == "" {
		return errors.New("member needs -name, -email and -password")
	}
	first, last := name, "-"
	if i := strings.LastIndex(name, " "); i > 0 {
		first, last = name[:i], name[i+1:]
	}
	dev := &mail.Dev{Quiet: true} // the invitation is read here, never sent
	au := auth.New(a)
	a.Mail = dev
	ps := people.New(a, au, approvals.New(a))

	var org, ownerID, ownerAcct uuid.UUID
	var ownerName string
	if err := a.DB.Tx(ctx, func(tx pgx.Tx) error {
		if slug != "" {
			return tx.QueryRow(ctx, `select id from hub.orgs where slug=$1`, slug).Scan(&org)
		}
		return tx.QueryRow(ctx, `select id from hub.orgs order by created_at limit 1`).Scan(&org)
	}); err != nil {
		return fmt.Errorf("organization not found: %w", err)
	}
	if err := a.DB.WithOrg(ctx, org, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `select p.id, p.account_id, p.first_name || ' ' || p.last_name from hub.people p join hub.person_roles pr on pr.person_id=p.id join hub.roles r on r.id=pr.role_id
			where r.locked and p.status='active' and p.account_id is not null order by p.created_at limit 1`).Scan(&ownerID, &ownerAcct, &ownerName)
	}); err != nil {
		return fmt.Errorf("this organization has no owner to add people: %w", err)
	}
	pr := &tenant.Principal{Org: org, Person: ownerID, Account: ownerAcct, Name: ownerName}

	// 1. the account, active, with the chosen password
	var acct uuid.UUID
	idx := crypto.Blind(cfg.BlindKey, "email", email)
	_ = a.DB.Tx(ctx, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `select id from hub.accounts where email_idx=$1`, idx).Scan(&acct)
	})
	if acct == uuid.Nil {
		var err error
		acct, err = au.Signup(ctx, name, email, "Tmp-Passw0rd-1", []auth.QuizIn{{Question: "q1", Answer: "aa"}, {Question: "q2", Answer: "bb"}, {Question: "q3", Answer: "cc"}}, "")
		if err != nil {
			return err
		}
	}
	if err := a.DB.Tx(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `update hub.accounts set pass_hash=$2, status='active', failed_attempts=0, locked_until=null, must_change=false, password_changed_at=now(), ver=ver+1, mac=hub.pending_mac() where id=$1`, acct, crypto.HashPassword(password, a.PW)); err != nil {
			return err
		}
		return a.Sign.Seal(ctx, tx, "accounts", "id=$1", acct)
	}); err != nil {
		return err
	}

	// 2. the person, joined to the account (skipped when the account already belongs to the organization)
	var joined bool
	_ = a.DB.WithOrg(ctx, org, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `select exists (select 1 from hub.people where account_id=$1)`, acct).Scan(&joined)
	})
	if !joined {
		res, err := ps.Onboard(ctx, pr, []people.NewPerson{{First: first, Last: last, Email: email, Title: "Staff member", Phone: "+254700000000", DOB: "1995-06-15", Gender: "Prefer not to say", Role: roleKey}})
		if err != nil {
			return err
		}
		if len(res) == 0 || res[0].Error != "" {
			return fmt.Errorf("onboarding: %+v", res)
		}
		msgs := dev.To(email)
		var tok []string
		if len(msgs) > 0 {
			tok = regexp.MustCompile(`t=([0-9a-f-]+\.[0-9a-f-]+\.[A-Za-z0-9_-]+)`).FindStringSubmatch(msgs[0].Body)
		}
		if tok == nil {
			return errors.New("no invitation link was made")
		}
		if _, err := ps.Accept(ctx, acct, tok[1]); err != nil {
			return err
		}
	}

	// 3. the development device: the same keys the browser derives from the email, so chats work before the first sign-in
	mac := hmac.New(sha256.New, []byte("hub-demo-seed"))
	mac.Write([]byte("hub-demo-device|" + email))
	d, err := e2ee.DeviceFromSeed(mac.Sum(nil))
	if err != nil {
		return err
	}
	identity, signing := d.Agree.PublicKey().Bytes(), d.Sign.Public().(ed25519.PublicKey)
	have := false
	_ = a.DB.Tx(ctx, func(tx pgx.Tx) error {
		var n int
		if err := tx.QueryRow(ctx, `select count(*) from hub.devices where account_id=$1 and identity_key=$2 and signing_key=$3 and revoked_at is null`, acct, identity, []byte(signing)).Scan(&n); err != nil {
			return err
		}
		have = n > 0
		return nil
	})
	if !have {
		sig := ed25519.Sign(d.Sign, keys.BindingMessage(identity, acct))
		if _, err := keys.New(a).Register(ctx, &tenant.Principal{Account: acct}, "Development device",
			base64.StdEncoding.EncodeToString(identity), base64.StdEncoding.EncodeToString(signing), base64.StdEncoding.EncodeToString(sig)); err != nil {
			return err
		}
	}
	_ = bytes.Equal
	fmt.Printf("%s is ready: active account, member of the organization, development device registered.\n", email)
	return nil
}

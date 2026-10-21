package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/config"
	"hub/internal/crypto"
	"hub/internal/integrity"
	"hub/internal/modules/auth"
	"hub/internal/modules/org"
)

// owner makes sure the person who runs the deployment exists, is active, and owns an organization. It is what a fresh go-live
// runs after the migrations, so there is someone able to sign in and set everything else up.
//
// The details come from flags or the environment (HUB_OWNER_NAME, HUB_OWNER_EMAIL, HUB_OWNER_PASSWORD, HUB_ORG_NAME, HUB_ORG_SLUG),
// never from the code. Running it again is safe: it activates and unlocks the account, and never resets a password unless
// -reset-password is given. A password it generates is printed once and must be changed at first sign-in.
func owner(ctx context.Context, a *app.App, cfg *config.Config, name, email, password, orgName, slug string, resetPassword bool) error {
	first := func(vals ...string) string {
		for _, v := range vals {
			if strings.TrimSpace(v) != "" {
				return strings.TrimSpace(v)
			}
		}
		return ""
	}
	name = first(name, os.Getenv("HUB_OWNER_NAME"))
	email = first(email, os.Getenv("HUB_OWNER_EMAIL"))
	password = first(password, os.Getenv("HUB_OWNER_PASSWORD"))
	orgName = first(orgName, os.Getenv("HUB_ORG_NAME"))
	slug = strings.ToLower(first(slug, os.Getenv("HUB_ORG_SLUG")))
	if name == "" || email == "" || orgName == "" || slug == "" {
		return fmt.Errorf("owner needs a name, an email, an organization name and a web name (flags -name -email -org -slug, or HUB_OWNER_NAME, HUB_OWNER_EMAIL, HUB_ORG_NAME, HUB_ORG_SLUG)")
	}
	norm, err := auth.NormEmail(email)
	if err != nil {
		return err
	}
	au := auth.New(a)
	idx := crypto.Blind(cfg.BlindKey, "email", norm)

	var acct uuid.UUID
	var status string
	err = a.DB.Tx(ctx, func(tx pgx.Tx) error {
		e := tx.QueryRow(ctx, `select id, status from hub.accounts where email_idx=$1`, idx).Scan(&acct, &status)
		if e == pgx.ErrNoRows {
			return nil
		}
		return e
	})
	if err != nil {
		return err
	}
	generated := false
	if acct == uuid.Nil {
		if password == "" {
			generated = true
			password = readablePassword()
		}
		// Account recovery needs three security questions; the owner sets real ones in their profile. Until then the answers
		// are random and unknown, so recovery by questions cannot be used by anyone else.
		rnd := func() string {
			b := make([]byte, 9)
			_, _ = rand.Read(b)
			return base64.RawURLEncoding.EncodeToString(b)
		}
		quiz := []auth.QuizIn{{Question: "Set in profile 1", Answer: rnd()}, {Question: "Set in profile 2", Answer: rnd()}, {Question: "Set in profile 3", Answer: rnd()}}
		acct, err = au.Signup(ctx, name, norm, password, quiz, "127.0.0.1")
		if err != nil {
			return fmt.Errorf("create the account: %w", err)
		}
		fmt.Println("account created:", norm)
	} else {
		fmt.Println("account exists:", norm)
	}
	// activated: active, not locked, failed attempts cleared (and a password reset only when asked)
	err = a.DB.Tx(ctx, func(tx pgx.Tx) error {
		if _, e := tx.Exec(ctx, `update hub.accounts set status='active', failed_attempts=0, locked_until=null, `+integrity.Bump+` where id=$1`, acct); e != nil {
			return e
		}
		return a.Sign.Seal(ctx, tx, "accounts", "id=$1", acct)
	})
	if err != nil {
		return err
	}
	if generated || resetPassword {
		if resetPassword && !generated {
			if password == "" {
				password = readablePassword()
				generated = true
			}
			if err := auth.CheckPassword(password); err != nil {
				return err
			}
			err = a.DB.Tx(ctx, func(tx pgx.Tx) error {
				if _, e := tx.Exec(ctx, `update hub.accounts set pass_hash=$2, password_changed_at=now(), `+integrity.Bump+` where id=$1`, acct, crypto.HashPassword(password, a.PW)); e != nil {
					return e
				}
				return a.Sign.Seal(ctx, tx, "accounts", "id=$1", acct)
			})
			if err != nil {
				return err
			}
		}
		// a password this tool chose (or was handed through the environment) must be replaced at the first sign-in
		err = a.DB.Tx(ctx, func(tx pgx.Tx) error {
			if _, e := tx.Exec(ctx, `update hub.accounts set must_change=true, `+integrity.Bump+` where id=$1`, acct); e != nil {
				return e
			}
			return a.Sign.Seal(ctx, tx, "accounts", "id=$1", acct)
		})
		if err != nil {
			return err
		}
	}
	// the organization: created with this person as its owner, or — if it exists — this person must already belong to it
	var orgID uuid.UUID
	err = a.DB.Tx(ctx, func(tx pgx.Tx) error {
		e := tx.QueryRow(ctx, `select id from hub.orgs where slug=$1`, slug).Scan(&orgID)
		if e == pgx.ErrNoRows {
			return nil
		}
		return e
	})
	if err != nil {
		return err
	}
	if orgID == uuid.Nil {
		c, err := org.New(a, au).Create(ctx, acct, name, orgName, slug, "127.0.0.1")
		if err != nil {
			return fmt.Errorf("create the organization: %w", err)
		}
		fmt.Printf("organization created: %s (%s), you are its owner\n", c.Name, c.Slug)
	} else {
		var member bool
		if err := a.DB.WithOrg(ctx, orgID, func(tx pgx.Tx) error {
			if _, e := tx.Exec(ctx, `update hub.people set status='active', `+integrity.Bump+` where account_id=$1 and status in ('inactive','invited')`, acct); e != nil {
				return e
			}
			if e := a.Sign.Seal(ctx, tx, "people", "account_id=$1", acct); e != nil {
				return e
			}
			return tx.QueryRow(ctx, `select exists (select 1 from hub.people p join hub.person_roles pr on pr.person_id=p.id join hub.roles r on r.id=pr.role_id where p.account_id=$1 and p.status='active' and r.key='owner')`, acct).Scan(&member)
		}); err != nil {
			return err
		}
		if !member {
			return fmt.Errorf("organization %q exists but %s is not its active owner; nothing was changed there", slug, norm)
		}
		fmt.Printf("organization exists: %s, you are its active owner\n", slug)
	}
	fmt.Println("activated: the account is active and unlocked.")
	if generated {
		fmt.Printf("\nFirst sign-in password (shown once, must be changed at first sign-in):\n  %s\n", password)
	}
	return nil
}

// readablePassword makes a first sign-in password that is easy to copy by eye: no letters that look alike (I l 1 O 0), in
// groups of four. It has an upper and a lower case letter and a number, as the password rules need.
func readablePassword() string {
	const letters = "ABCDEFGHJKLMNPQRSTUVWXYZ"
	const lower = "abcdefghijkmnpqrstuvwxyz"
	const digits = "23456789"
	pick := func(set string, n int) string {
		b := make([]byte, n)
		_, _ = rand.Read(b)
		out := make([]byte, n)
		for i := range b {
			out[i] = set[int(b[i])%len(set)]
		}
		return string(out)
	}
	return pick(letters, 4) + "-" + pick(lower, 4) + "-" + pick(digits, 4) + "-" + pick(lower+digits, 4)
}

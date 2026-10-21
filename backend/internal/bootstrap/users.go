package bootstrap

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/access"
	"hub/internal/app"
	"hub/internal/config"
	"hub/internal/crypto"
	"hub/internal/integrity"
	"hub/internal/modules/auth"
	"hub/internal/modules/chat"
	pp "hub/internal/people"
)

// SeedUser is one account that is kept in place on every start: it exists, is active and unlocked, can sign in with this password,
// and is a member of the organization with this role, so it can take part in the organization's forums, wall and chats.
type SeedUser struct{ Email, Name, Password, Role string }

// ParseSeedUsers reads HUB_SEED_USERS: entries separated by ";", each "email|name|password|role" (name and role optional; the
// role defaults to staff).
func ParseSeedUsers(s string) ([]SeedUser, error) {
	var out []SeedUser
	for _, e := range strings.Split(s, ";") {
		if e = strings.TrimSpace(e); e == "" {
			continue
		}
		f := strings.Split(e, "|")
		if len(f) < 3 {
			return nil, fmt.Errorf("HUB_SEED_USERS: each entry is email|name|password|role, got %q", e)
		}
		u := SeedUser{Email: strings.TrimSpace(f[0]), Name: strings.TrimSpace(f[1]), Password: f[2], Role: "staff"}
		if len(f) > 3 && strings.TrimSpace(f[3]) != "" {
			u.Role = strings.ToLower(strings.TrimSpace(f[3]))
		}
		if u.Name == "" {
			u.Name = strings.Split(u.Email, "@")[0]
		}
		out = append(out, u)
	}
	return out, nil
}

// EnsureUsers makes every seeded user exist, be active and be a member of the organization. Running it again changes nothing
// except that each password is set back to the configured one.
func EnsureUsers(ctx context.Context, a *app.App, cfg *config.Config, slug string, users []SeedUser) error {
	var org uuid.UUID
	if err := a.DB.Tx(ctx, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `select id from hub.orgs where slug=$1`, slug).Scan(&org)
	}); err != nil {
		return fmt.Errorf("seed users: no organization %q yet", slug)
	}
	au := auth.New(a)
	for _, u := range users {
		if err := ensureUser(ctx, a, cfg, au, org, u); err != nil {
			return fmt.Errorf("seed user %s: %w", u.Email, err)
		}
		slog.Info("seeded user ready", "email", u.Email, "role", u.Role)
	}
	return nil
}

func ensureUser(ctx context.Context, a *app.App, cfg *config.Config, au *auth.Service, org uuid.UUID, u SeedUser) error {
	norm, err := auth.NormEmail(u.Email)
	if err != nil {
		return err
	}
	if err := auth.CheckPassword(u.Password); err != nil {
		return err
	}
	idx := crypto.Blind(cfg.BlindKey, "email", norm)
	var acct uuid.UUID
	err = a.DB.Tx(ctx, func(tx pgx.Tx) error {
		e := tx.QueryRow(ctx, `select id from hub.accounts where email_idx=$1`, idx).Scan(&acct)
		if e == pgx.ErrNoRows {
			return nil
		}
		return e
	})
	if err != nil {
		return err
	}
	if acct == uuid.Nil {
		rnd := func() string {
			b := make([]byte, 9)
			_, _ = rand.Read(b)
			return base64.RawURLEncoding.EncodeToString(b)
		}
		quiz := []auth.QuizIn{{Question: "Set in profile 1", Answer: rnd()}, {Question: "Set in profile 2", Answer: rnd()}, {Question: "Set in profile 3", Answer: rnd()}}
		if acct, err = au.Signup(ctx, u.Name, norm, u.Password, quiz, "127.0.0.1"); err != nil {
			return err
		}
	}
	// active, unlocked, with the configured password and no forced change
	if err := a.DB.Tx(ctx, func(tx pgx.Tx) error {
		if _, e := tx.Exec(ctx, `update hub.accounts set status='active', failed_attempts=0, locked_until=null, must_change=false, pass_hash=$2, password_changed_at=now(), `+integrity.Bump+` where id=$1`, acct, crypto.HashPassword(u.Password, a.PW)); e != nil {
			return e
		}
		return a.Sign.Seal(ctx, tx, "accounts", "id=$1", acct)
	}); err != nil {
		return err
	}
	// a member of the organization, active, with a role
	return a.DB.WithOrg(ctx, org, func(tx pgx.Tx) error {
		seal := func(table, where string, args ...any) error { return a.Sign.Seal(ctx, tx, table, where, args...) }
		emailIdx := pp.EmailIdx(cfg.BlindKey, org, norm)
		var person uuid.UUID
		e := tx.QueryRow(ctx, `select id from hub.people where org_id=$1 and email_idx=$2 and status <> 'deleted'`, org, emailIdx).Scan(&person)
		if e != nil && e != pgx.ErrNoRows {
			return e
		}
		var role uuid.UUID
		roleErr := tx.QueryRow(ctx, `select id from hub.roles where key=$1`, u.Role).Scan(&role)
		if roleErr == pgx.ErrNoRows && u.Role == "staff" {
			if _, err := access.InstallStarterRoles(ctx, tx, org, app.NewID, seal); err != nil { // the starter roles include Staff
				return err
			}
			roleErr = tx.QueryRow(ctx, `select id from hub.roles where key=$1`, u.Role).Scan(&role)
		}
		if roleErr != nil {
			return fmt.Errorf("the organization has no role %q", u.Role)
		}
		if person == uuid.Nil {
			person = app.NewID()
			dek, _, err := a.OrgDEK(ctx, tx, org)
			if err != nil {
				return err
			}
			first, last := splitName(u.Name)
			emailEnc, err := pp.Enc(dek, org, person, pp.ColEmail, norm)
			if err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `insert into hub.people (id, org_id, account_id, first_name, last_name, title, email_idx, email_enc, joined_on) values ($1,$2,$3,$4,$5,'Member',$6,$7,current_date)`,
				person, org, acct, first, last, emailIdx, emailEnc); err != nil {
				return err
			}
			if err := seal("people", "id=$1", person); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `insert into hub.person_roles (org_id, person_id, role_id) values ($1,$2,$3)`, org, person, role); err != nil {
				return err
			}
			if err := seal("person_roles", "org_id=$1", org); err != nil {
				return err
			}
		} else {
			if _, err := tx.Exec(ctx, `update hub.people set status='active', account_id=$2, `+integrity.Bump+` where id=$1`, person, acct); err != nil {
				return err
			}
			if err := seal("people", "id=$1", person); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(ctx, `insert into hub.account_orgs (account_id, org_id) values ($1,$2) on conflict do nothing`, acct, org); err != nil {
			return err
		}
		if err := seal("account_orgs", "account_id=$1 and org_id=$2", acct, org); err != nil {
			return err
		}
		return chat.SyncAudience(ctx, tx, a.Sign, org, acct)
	})
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

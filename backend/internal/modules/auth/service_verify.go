package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pquerna/otp/totp"

	"hub/internal/app"
	"hub/internal/crypto"
	"hub/internal/httpx"
	"hub/internal/integrity"
	hubmail "hub/internal/mail"
)

// Verify checks an access token and that its session has not been revoked.
func (s *Service) Verify(ctx context.Context, token string) (*Claims, error) {
	var c Claims
	_, err := jwt.ParseWithClaims(token, &c, func(*jwt.Token) (any, error) { return s.A.Cfg.TokenKey, nil },
		jwt.WithValidMethods([]string{"HS256"}), jwt.WithIssuer("hub"), jwt.WithExpirationRequired(), jwt.WithLeeway(5*time.Second))
	if err != nil {
		return nil, httpx.ErrUnauthorized
	}
	if why, err := s.A.Rdb.Get(ctx, "revoked:"+c.Sid).Result(); err == nil {
		return nil, endedErr(why)
	}
	return &c, nil
}

func (s *Service) markRevoked(ctx context.Context, ids []uuid.UUID) {
	for _, id := range ids {
		s.A.Rdb.Set(ctx, "revoked:"+id.String(), 1, s.A.Cfg.AccessTTL+time.Minute)
	}
}

// ---- refresh / logout ----

func (s *Service) Refresh(ctx context.Context, refresh, orgSlug, ip, ua string) (*Tokens, error) {
	parts := strings.SplitN(refresh, ".", 2)
	sid, err := uuid.Parse(parts[0])
	if err != nil || len(parts) != 2 {
		return nil, httpx.ErrUnauthorized
	}
	var tok *Tokens
	var killed, idle []uuid.UUID
	var renewedFrom uuid.UUID
	var outErr error
	err = s.A.DB.Tx(ctx, func(tx pgx.Tx) error {
		x, err := scanSession(tx.QueryRow(ctx, `select `+sessionCols+` from hub.sessions where id=$1 for update`, sid))
		if errors.Is(err, pgx.ErrNoRows) {
			outErr = httpx.ErrUnauthorized
			return nil
		}
		if err != nil {
			return err
		}
		if !crypto.Equal(x.RefreshHash, hashSecret(parts[1])) {
			outErr = httpx.ErrUnauthorized
			return nil
		}
		if x.RevokedAt != nil {
			// signed out on purpose (a sign-in elsewhere, or inactivity): say so
			if e := s.whyEnded(ctx, x.ID); e != httpx.ErrUnauthorized {
				outErr = e
				return nil
			}
			// A rotated or revoked token came back: assume theft and end the whole family.
			killed, err = s.revokeFamily(ctx, tx, x.Family)
			outErr = httpx.ErrUnauthorized
			return err
		}
		// inactive past the organization's limit: the session ends instead of being renewed
		if x.Org != nil && s.idleExpired(ctx, *x.Org, x.ID) {
			if err := s.revokeSession(ctx, tx, x); err != nil {
				return err
			}
			idle = append(idle, x.ID)
			outErr = errIdle
			return nil
		}
		if x.ExpiresAt.Before(time.Now()) {
			outErr = httpx.ErrUnauthorized
			return nil
		}
		a, err := s.accountByID(ctx, tx, x.Account, false)
		if err != nil || a.Status != "active" {
			outErr = httpx.ErrUnauthorized
			return nil
		}
		if err := s.revokeSession(ctx, tx, x); err != nil {
			return err
		}
		if orgSlug == "" && x.Org != nil {
			_ = tx.QueryRow(ctx, `select slug from hub.orgs where id=$1`, *x.Org).Scan(&orgSlug)
		}
		tok, err = s.issue(ctx, tx, a, orgSlug, x.Family, ip, ua)
		killed = append(killed, x.ID) // the old access token dies with the old session
		renewedFrom = x.ID
		return err
	})
	s.markRevoked(ctx, killed)
	s.markEnded(ctx, idle, whyIdle)
	// a renewed session keeps the last activity of the one it replaces (renewing is not activity)
	if tok != nil {
		if nsid, e := uuid.Parse(strings.SplitN(tok.Refresh, ".", 2)[0]); e == nil {
			if v, e := s.A.Rdb.Get(ctx, actKey(renewedFrom)).Result(); e == nil {
				s.A.Rdb.Set(ctx, actKey(nsid), v, 25*time.Hour)
			}
		}
	}
	if err != nil {
		return nil, err
	}
	if outErr != nil {
		return nil, outErr
	}
	return tok, nil
}

func (s *Service) Logout(ctx context.Context, sid uuid.UUID) error {
	var ids []uuid.UUID
	err := s.A.DB.Tx(ctx, func(tx pgx.Tx) error {
		x, err := scanSession(tx.QueryRow(ctx, `select `+sessionCols+` from hub.sessions where id=$1 for update`, sid))
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		if x.RevokedAt == nil {
			ids = append(ids, x.ID)
			return s.revokeSession(ctx, tx, x)
		}
		return nil
	})
	s.markRevoked(ctx, ids)
	return err
}

// ---- change password (signed in) ----

func (s *Service) ChangePassword(ctx context.Context, acct, keepSession uuid.UUID, current, next, ip string) error {
	if err := CheckPassword(next); err != nil {
		return err
	}
	var killed []uuid.UUID
	var outErr error
	err := s.A.DB.Tx(ctx, func(tx pgx.Tx) error {
		a, err := s.accountByID(ctx, tx, acct, true)
		if err != nil {
			return err
		}
		if !crypto.VerifyPassword(current, a.PassHash) {
			outErr = httpx.Err(400, "wrong_password", "Your current password is not right.")
			return nil
		}
		if crypto.VerifyPassword(next, a.PassHash) {
			outErr = httpx.Bad("Choose a password you have not just used.")
			return nil
		}
		a.PassHash, a.PasswordChangedAt, a.MustChange = crypto.HashPassword(next, s.A.PW), app.Now(), false
		if err := s.saveAccount(ctx, tx, a); err != nil {
			return err
		}
		// every other device is signed out; this one stays
		all, err := s.revokeAllFor(ctx, tx, acct, keepSession)
		if err != nil {
			return err
		}
		for _, id := range all {
			if id != keepSession {
				killed = append(killed, id)
			}
		}
		return nil
	})
	s.markRevoked(ctx, killed)
	if err != nil {
		return err
	}
	return outErr
}

// ---- two-step sign-in (TOTP) ----

func (s *Service) MFASetup(ctx context.Context, acct uuid.UUID) (secret, uri string, err error) {
	var email string
	err = s.A.DB.Tx(ctx, func(tx pgx.Tx) error {
		a, err := s.accountByID(ctx, tx, acct, true)
		if err != nil {
			return err
		}
		email, _ = s.decEmail(a)
		key, err := totp.Generate(totp.GenerateOpts{Issuer: "Hub", AccountName: email})
		if err != nil {
			return err
		}
		secret, uri = key.Secret(), key.URL()
		a.MFASecretEnc, err = crypto.Field(s.A.PlatformKey(), crypto.Context{Org: "platform", Table: "accounts", Column: "mfa_secret_enc", RowID: a.ID.String()}, []byte(secret))
		if err != nil {
			return err
		}
		a.MFAEnabled = false // turns on only after a code is confirmed
		return s.saveAccount(ctx, tx, a)
	})
	return
}

func (s *Service) MFAConfirm(ctx context.Context, acct uuid.UUID, code string) error {
	return s.A.DB.Tx(ctx, func(tx pgx.Tx) error {
		a, err := s.accountByID(ctx, tx, acct, true)
		if err != nil {
			return err
		}
		if a.MFASecretEnc == nil {
			return httpx.Bad("Start two-step setup first.")
		}
		sec, err := crypto.FieldOpen(s.A.PlatformKey(), crypto.Context{Org: "platform", Table: "accounts", Column: "mfa_secret_enc", RowID: a.ID.String()}, a.MFASecretEnc)
		if err != nil {
			return err
		}
		if !totp.Validate(code, string(sec)) {
			return httpx.Bad("That code is not right.")
		}
		a.MFAEnabled = true
		return s.saveAccount(ctx, tx, a)
	})
}

var _ = fmt.Sprint

// AccountEmail returns the account's email (decrypted). Other modules use it when they create the account's person record.
func (s *Service) AccountEmail(ctx context.Context, tx pgx.Tx, id uuid.UUID) (string, error) {
	a, err := s.accountByID(ctx, tx, id, false)
	if err != nil {
		return "", err
	}
	return s.decEmail(a)
}

// RevokeAccountInOrg signs the account out of one organization on every device (used when a person is blocked,
// deactivated or deleted). Their sessions in other organizations are untouched.
func (s *Service) RevokeAccountInOrg(ctx context.Context, account, org uuid.UUID) {
	var killed []uuid.UUID
	_ = s.A.DB.Tx(ctx, func(tx pgx.Tx) error {
		rs, err := tx.Query(ctx, `select `+sessionCols+` from hub.sessions where account_id=$1 and org_id=$2 and revoked_at is null for update`, account, org)
		if err != nil {
			return err
		}
		var live []Session
		for rs.Next() {
			x, err := scanSession(rs)
			if err != nil {
				rs.Close()
				return err
			}
			live = append(live, x)
		}
		rs.Close()
		for _, x := range live {
			if err := s.revokeSession(ctx, tx, x); err != nil {
				return err
			}
			killed = append(killed, x.ID)
		}
		return nil
	})
	s.markRevoked(ctx, killed)
}

// AdminSetTemporaryPassword replaces the account's password with a random one that must be changed at next sign-in,
// signs the account out everywhere, and tells the owner by email. The temporary password is returned once.
func (s *Service) AdminSetTemporaryPassword(ctx context.Context, account uuid.UUID) (string, error) {
	temp := crypto.RandomToken(9) + "aA1"
	var killed []uuid.UUID
	var notify string
	err := s.A.DB.Tx(ctx, func(tx pgx.Tx) error {
		a, err := s.accountByID(ctx, tx, account, true)
		if err != nil {
			return err
		}
		a.PassHash, a.PasswordChangedAt, a.MustChange, a.FailedAttempts, a.LockedUntil = crypto.HashPassword(temp, s.A.PW), app.Now(), true, 0, nil
		if err := s.saveAccount(ctx, tx, a); err != nil {
			return err
		}
		if killed, err = s.revokeAllFor(ctx, tx, account, uuid.Nil); err != nil {
			return err
		}
		notify, _ = s.decEmail(a)
		return nil
	})
	s.markRevoked(ctx, killed)
	if err != nil {
		return "", err
	}
	if notify != "" {
		_ = hubmail.SendCard(ctx, s.A.Mail, notify, "Your Hub password was reset", hubmail.Card{Accent: hubmail.Amber, Title: "An administrator reset your password", Preheader: "Ask them for your temporary password",
			Lead: "Ask your administrator for the temporary password. You will choose a new password when you sign in."})
	}
	return temp, nil
}

// How long a sign-in block lasts. It lifts by itself; the record of it stays on the Security page.
const blockFor = 2 * time.Minute

// How many sign-in attempts one address gets per window before it is blocked.
const maxTries = 5

// recordBlock keeps a record of a sign-in block (once while it lasts). The account is found from the address when it belongs
// to someone; otherwise the record shows only a masked address, because nobody in Hub owns it.
func (s *Service) recordBlock(ctx context.Context, kind string, idx []byte, email, ip, reason string, wait time.Duration) {
	if wait <= 0 {
		wait = blockFor
	}
	hint := email
	if i := strings.Index(email, "@"); i > 1 {
		hint = email[:1] + "***" + email[i:]
	}
	_ = s.A.DB.Tx(ctx, func(tx pgx.Tx) error {
		var acct *uuid.UUID
		if a, err := s.accountByIdx(ctx, tx, idx, false); err == nil {
			acct = &a.ID
		}
		var n int
		if err := tx.QueryRow(ctx, `select count(*) from hub.login_blocks where kind=$1 and until > now() and ip=$2 and account_id is not distinct from $3`, kind, ip, acct).Scan(&n); err != nil || n > 0 {
			return err
		}
		_, err := tx.Exec(ctx, `insert into hub.login_blocks (kind, account_id, email_hint, ip, reason, until) values ($1,$2,$3,$4,$5,$6)`, kind, acct, hint, ip, reason, time.Now().Add(wait))
		return err
	})
}

// mailLocked tells the owner of an account that it was locked after wrong passwords.
func (s *Service) mailLocked(email, name, ip string, until time.Time) {
	loc := ""
	if s.A.Geo != nil && ip != "" {
		loc = s.A.Geo(ip)
	}
	rows := [][2]string{{"Account", email}, {"Attempts from", ip}}
	if loc != "" {
		rows = append(rows, [2]string{"Location", loc})
	}
	rows = append(rows, [2]string{"Locked until", until.UTC().Format("15:04 UTC")}, [2]string{"Unlocks", "by itself, nothing to do"})
	_ = hubmail.SendCard(context.Background(), s.A.Mail, email, "Your Hub account was temporarily locked", hubmail.Card{Greeting: "Hello " + firstWord(name) + ",", Title: "Your account was temporarily locked", Preheader: "Five wrong passwords were entered",
		Lead: "Someone entered the wrong password five times in a row, so Hub locked your account for a few minutes to keep it safe. It unlocks by itself.", Rows: rows,
		Note: "If this was you, wait a few minutes and try again, or reset your password. If it was not you, reset your password now and tell your administrator."})
}

// mailNewAddress tells the owner of an account that someone signed in from a browser or device it had not used before.
func (s *Service) mailNewAddress(email, name, ip, ua string) {
	loc := ""
	if s.A.Geo != nil {
		loc = s.A.Geo(ip)
	}
	if len(ua) > 140 {
		ua = ua[:140]
	}
	rows := [][2]string{{"When", time.Now().UTC().Format("2 Jan 2006, 15:04 UTC")}, {"Address", ip}}
	if loc != "" {
		rows = append(rows, [2]string{"Location", loc})
	}
	if ua != "" {
		rows = append(rows, [2]string{"Browser", ua})
	}
	_ = hubmail.SendCard(context.Background(), s.A.Mail, email, "New device signed in to your Hub account", hubmail.Card{Greeting: "Hello " + firstWord(name) + ",", Title: "A new device signed in", Preheader: "A sign-in from a device you have not used before",
		Lead: "Your account was just signed in to from a browser or device it has not used before. Signing in here signed you out of your other browsers and devices.", Rows: rows,
		Note: "If this was you, no action is needed. If it was not you, reset your password now and tell your administrator."})
}

func firstWord(n string) string {
	if i := strings.IndexByte(n, ' '); i > 0 {
		return n[:i]
	}
	if n == "" {
		return "there"
	}
	return n
}

// ChangeAccountEmail changes the address an account signs in with (an administrator editing a person). It returns the old address
// so the caller can tell its owner. An address another account already uses is refused.
func (s *Service) ChangeAccountEmail(ctx context.Context, tx pgx.Tx, account uuid.UUID, email string) (string, error) {
	e, err := NormEmail(email)
	if err != nil {
		return "", err
	}
	if !s.domainAllowed(e) {
		return "", errDomain
	}
	a, err := s.accountByID(ctx, tx, account, true)
	if err != nil {
		return "", err
	}
	old, err := s.decEmail(a)
	if err != nil {
		return "", err
	}
	if old == e {
		return old, nil
	}
	ct, err := s.encEmail(account, e)
	if err != nil {
		return "", err
	}
	if _, err := tx.Exec(ctx, "savepoint email_change"); err != nil {
		return "", err
	}
	if _, err := tx.Exec(ctx, `update hub.accounts set email_idx=$2, email_enc=$3, updated_at=now(), `+integrity.Bump+` where id=$1`, account, s.emailIdx(e), ct); err != nil {
		if strings.Contains(err.Error(), "accounts_email_idx_key") {
			_, _ = tx.Exec(ctx, "rollback to savepoint email_change")
			return "", httpx.Err(409, "email_in_use", "Another account already uses this email address.")
		}
		return "", err
	}
	if err := s.A.Sign.Seal(ctx, tx, "accounts", "id=$1", account); err != nil {
		return "", err
	}
	return old, nil
}

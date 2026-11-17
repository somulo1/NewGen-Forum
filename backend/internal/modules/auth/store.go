// Package auth is the accounts, sessions and password-recovery module.
package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/httpx"
	"hub/internal/integrity"
)

type Account struct {
	ID                               uuid.UUID
	EmailIdx, EmailEnc, MFASecretEnc []byte
	DisplayName, PassHash, Status    string
	MFAEnabled                       bool
	FailedAttempts                   int
	LockedUntil                      *time.Time
	PasswordChangedAt                time.Time
	MustChange                       bool
}

type Quiz struct {
	Account    uuid.UUID
	Position   int
	Question   string
	AnswerHash string
}

type Session struct {
	ID, Account, Family uuid.UUID
	Org, Device         *uuid.UUID
	RefreshHash         []byte
	ExpiresAt           time.Time
	RevokedAt           *time.Time
	Browser             []byte // hash of the browser's own token, on sign-in
}

type Reset struct {
	ID                             uuid.UUID
	Account                        *uuid.UUID
	EmailIdx                       []byte
	Stage                          string
	OTPHash                        []byte
	OTPExpires                     time.Time
	OTPTries, QuizTries, LinkTries int
	LinkHash, LinkOTPHash          []byte
	LinkExpires                    *time.Time
}

const accountCols = `id, email_idx, email_enc, mfa_secret_enc, display_name, pass_hash, status, mfa_enabled, failed_attempts, locked_until, password_changed_at, must_change`

func scanAccount(r pgx.Row) (Account, error) {
	var a Account
	err := r.Scan(&a.ID, &a.EmailIdx, &a.EmailEnc, &a.MFASecretEnc, &a.DisplayName, &a.PassHash, &a.Status, &a.MFAEnabled, &a.FailedAttempts, &a.LockedUntil, &a.PasswordChangedAt, &a.MustChange)
	a.PasswordChangedAt = a.PasswordChangedAt.UTC()
	return a, err
}

func (s *Service) accountByIdx(ctx context.Context, tx pgx.Tx, idx []byte, lock bool) (Account, error) {
	q := "select " + accountCols + " from hub.accounts where email_idx=$1"
	if lock {
		q += " for update"
	}
	return scanAccount(tx.QueryRow(ctx, q, idx))
}

func (s *Service) accountByID(ctx context.Context, tx pgx.Tx, id uuid.UUID, lock bool) (Account, error) {
	q := "select " + accountCols + " from hub.accounts where id=$1"
	if lock {
		q += " for update"
	}
	return scanAccount(tx.QueryRow(ctx, q, id))
}

// saveAccount writes the account back; the row is bumped to the next version and sealed with a fresh MAC.
func (s *Service) saveAccount(ctx context.Context, tx pgx.Tx, a Account) error {
	_, err := tx.Exec(ctx, `update hub.accounts set mfa_secret_enc=$2, display_name=$3, pass_hash=$4, status=$5, mfa_enabled=$6, failed_attempts=$7, locked_until=$8, password_changed_at=$9, must_change=$10, updated_at=now(), `+integrity.Bump+` where id=$1`,
		a.ID, a.MFASecretEnc, a.DisplayName, a.PassHash, a.Status, a.MFAEnabled, a.FailedAttempts, a.LockedUntil, a.PasswordChangedAt, a.MustChange)
	if err != nil {
		return err
	}
	return s.A.Sign.Seal(ctx, tx, "accounts", "id=$1", a.ID)
}

func (s *Service) insertSession(ctx context.Context, tx pgx.Tx, x Session, ip, ua string) error {
	_, err := tx.Exec(ctx, `insert into hub.sessions (id, account_id, org_id, device_id, family_id, refresh_hash, expires_at, revoked_at, ip, user_agent, browser_hash) values ($1,$2,$3,$4,$5,$6,$7,$8,nullif($9,'')::inet,$10,$11)`,
		x.ID, x.Account, x.Org, x.Device, x.Family, x.RefreshHash, x.ExpiresAt, x.RevokedAt, ip, ua, x.Browser)
	if err != nil {
		return err
	}
	return s.A.Sign.Seal(ctx, tx, "sessions", "id=$1", x.ID)
}

func scanSession(r pgx.Row) (Session, error) {
	var x Session
	err := r.Scan(&x.ID, &x.Account, &x.Org, &x.Device, &x.Family, &x.RefreshHash, &x.ExpiresAt, &x.RevokedAt)
	x.ExpiresAt = x.ExpiresAt.UTC()
	if x.RevokedAt != nil {
		t := x.RevokedAt.UTC()
		x.RevokedAt = &t
	}
	return x, err
}

const sessionCols = `id, account_id, org_id, device_id, family_id, refresh_hash, expires_at, revoked_at`

func (s *Service) revokeSession(ctx context.Context, tx pgx.Tx, x Session) error {
	now := app.Now()
	x.RevokedAt = &now
	if _, err := tx.Exec(ctx, `update hub.sessions set revoked_at=$2, `+integrity.Bump+` where id=$1`, x.ID, now); err != nil {
		return err
	}
	return s.A.Sign.Seal(ctx, tx, "sessions", "id=$1", x.ID)
}

// revokeWhere revokes every live session the condition picks (written against hub.sessions s) and returns their ids.
func (s *Service) revokeWhere(ctx context.Context, tx pgx.Tx, where string, args ...any) ([]uuid.UUID, error) {
	rs, err := tx.Query(ctx, `select `+sessionCols+` from hub.sessions s where s.revoked_at is null and `+where+` for update`, args...)
	if err != nil {
		return nil, err
	}
	var live []Session
	for rs.Next() {
		x, err := scanSession(rs)
		if err != nil {
			rs.Close()
			return nil, err
		}
		live = append(live, x)
	}
	rs.Close()
	var ids []uuid.UUID
	for _, x := range live {
		if err := s.revokeSession(ctx, tx, x); err != nil {
			return nil, err
		}
		ids = append(ids, x.ID)
	}
	return ids, nil
}

// revokeFamily revokes every live session in a refresh-token family (reuse of a rotated token means it was stolen).
func (s *Service) revokeFamily(ctx context.Context, tx pgx.Tx, family uuid.UUID) ([]uuid.UUID, error) {
	return s.revokeWhere(ctx, tx, `s.family_id=$1`, family)
}

// revokeBrowsers revokes the account's browser sessions except `keep` (a password change or reset). The phone stays signed in:
// only confirming another phone, or the person being blocked, deactivated or removed, signs it out.
func (s *Service) revokeBrowsers(ctx context.Context, tx pgx.Tx, account, keep uuid.UUID) ([]uuid.UUID, error) {
	return s.revokeWhere(ctx, tx, `s.account_id=$1 and s.id<>$2 and not exists (select 1 from hub.session_devices d where d.family_id = s.family_id)`, account, keep)
}

// familyEnded says why a sign-in has no live session left (its newest session carries the reason).
func (s *Service) familyEnded(ctx context.Context, tx pgx.Tx, family uuid.UUID) error {
	var last uuid.UUID
	if err := tx.QueryRow(ctx, `select id from hub.sessions where family_id=$1 order by created_at desc limit 1`, family).Scan(&last); err != nil {
		return httpx.ErrUnauthorized
	}
	return s.whyEnded(ctx, last)
}

// revokeSlot revokes the account's live sessions of one kind: its phone (a sign-in bound to a phone key) or its browser.
// A person is signed in on at most one browser and one phone; a new sign-in replaces only the one of its own kind.
func (s *Service) revokeSlot(ctx context.Context, tx pgx.Tx, account uuid.UUID, phone bool) ([]uuid.UUID, error) {
	return s.revokeWhere(ctx, tx, `s.account_id=$1 and exists (select 1 from hub.session_devices d where d.family_id = s.family_id) = $2`, account, phone)
}

// Package auth is the accounts, sessions and password-recovery module.
package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
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
	_, err := tx.Exec(ctx, `insert into hub.sessions (id, account_id, org_id, device_id, family_id, refresh_hash, expires_at, revoked_at, ip, user_agent) values ($1,$2,$3,$4,$5,$6,$7,$8,nullif($9,'')::inet,$10)`,
		x.ID, x.Account, x.Org, x.Device, x.Family, x.RefreshHash, x.ExpiresAt, x.RevokedAt, ip, ua)
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

// revokeFamily revokes every live session in a refresh-token family (reuse of a rotated token means it was stolen).
func (s *Service) revokeFamily(ctx context.Context, tx pgx.Tx, family uuid.UUID) ([]uuid.UUID, error) {
	rs, err := tx.Query(ctx, `select `+sessionCols+` from hub.sessions where family_id=$1 and revoked_at is null for update`, family)
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

// revokeAllFor revokes every live session of the account except `keep` (pass uuid.Nil to revoke all).
func (s *Service) revokeAllFor(ctx context.Context, tx pgx.Tx, account, keep uuid.UUID) ([]uuid.UUID, error) {
	rs, err := tx.Query(ctx, `select `+sessionCols+` from hub.sessions where account_id=$1 and id<>$2 and revoked_at is null for update`, account, keep)
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

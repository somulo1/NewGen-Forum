package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/crypto"
	"hub/internal/httpx"
	"hub/internal/integrity"
	hubmail "hub/internal/mail"
)

// Password recovery. Five steps, and nothing in any response tells an outsider whether an account exists:
//  1. start     email in -> a code is emailed (only if the account exists); the reply is the same either way
//  2. code      the emailed 6-digit code (10 minutes, 5 tries)
//  3. quiz      the three security questions; all must be right (3 tries)
//  4. link      a single-use link and a second code are emailed (30 minutes)
//  5. complete  token + second code + a new password; every session is signed out

var (
	errCode = httpx.Err(400, "invalid_code", "Incorrect code.")
	errQuiz = httpx.Err(400, "invalid_answers", "Incorrect answers.")
	errLink = httpx.Err(400, "invalid_link", "This link is no longer valid.")
)

const resetCols = `id, account_id, email_idx, stage, otp_hash, otp_expires, otp_tries, quiz_tries, link_hash, link_otp_hash, link_expires, link_tries`

func scanReset(r pgx.Row) (Reset, error) {
	var x Reset
	err := r.Scan(&x.ID, &x.Account, &x.EmailIdx, &x.Stage, &x.OTPHash, &x.OTPExpires, &x.OTPTries, &x.QuizTries, &x.LinkHash, &x.LinkOTPHash, &x.LinkExpires, &x.LinkTries)
	x.OTPExpires = x.OTPExpires.UTC()
	if x.LinkExpires != nil {
		t := x.LinkExpires.UTC()
		x.LinkExpires = &t
	}
	return x, err
}

func (s *Service) saveReset(ctx context.Context, tx pgx.Tx, x Reset) error {
	_, err := tx.Exec(ctx, `update hub.password_resets set stage=$2, otp_tries=$3, quiz_tries=$4, link_hash=$5, link_otp_hash=$6, link_expires=$7, link_tries=$8, `+integrity.Bump+` where id=$1`,
		x.ID, x.Stage, x.OTPTries, x.QuizTries, x.LinkHash, x.LinkOTPHash, x.LinkExpires, x.LinkTries)
	if err != nil {
		return err
	}
	return s.A.Sign.Seal(ctx, tx, "password_resets", "id=$1", x.ID)
}

func (s *Service) codeMAC(kind string, id uuid.UUID, code string) []byte {
	return crypto.MAC(s.A.Cfg.TokenKey, []byte("hub/reset/"+kind), id[:], []byte(code))
}

func (s *Service) siteBase() string { return strings.TrimRight(s.A.Cfg.AllowedOrigins[0], "/") }

func (s *Service) ResetStart(ctx context.Context, email, ip string) (uuid.UUID, error) {
	e, err := NormEmail(email)
	if err != nil {
		return uuid.New(), nil // same reply as everything else: never tell a bad address from an unknown one
	}
	if ok, _ := s.A.Limit.Allow(ctx, "reset-ip", ip, 10, time.Hour, true); !ok {
		return uuid.Nil, errLocked
	}
	if ok, _ := s.A.Limit.Allow(ctx, "reset-mail", string(s.emailIdx(e)), 5, time.Hour, true); !ok {
		return uuid.New(), nil // quietly stop sending; the caller sees nothing different
	}
	id := app.NewID()
	var send func()
	err = s.A.DB.Tx(ctx, func(tx pgx.Tx) error {
		a, err := s.accountByIdx(ctx, tx, s.emailIdx(e), false)
		if errors.Is(err, pgx.ErrNoRows) {
			s.codeMAC("otp", id, crypto.Digits(6)) // do the same work as the real path
			return nil
		}
		if err != nil {
			return err
		}
		// older open attempts for this address are cancelled
		old, err := tx.Query(ctx, `select `+resetCols+` from hub.password_resets where email_idx=$1 and stage in ('otp','quiz','link') for update`, s.emailIdx(e))
		if err != nil {
			return err
		}
		var olds []Reset
		for old.Next() {
			x, err := scanReset(old)
			if err != nil {
				old.Close()
				return err
			}
			olds = append(olds, x)
		}
		old.Close()
		for _, x := range olds {
			x.Stage = "dead"
			if err := s.saveReset(ctx, tx, x); err != nil {
				return err
			}
		}
		code := crypto.Digits(6)
		r := Reset{ID: id, Account: &a.ID, EmailIdx: s.emailIdx(e), Stage: "otp", OTPHash: s.codeMAC("otp", id, code), OTPExpires: app.Now().Add(10 * time.Minute)}
		if _, err := tx.Exec(ctx, `insert into hub.password_resets (id, account_id, email_idx, stage, otp_hash, otp_expires, ip) values ($1,$2,$3,'otp',$4,$5,nullif($6,'')::inet)`,
			r.ID, r.Account, r.EmailIdx, r.OTPHash, r.OTPExpires, ip); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "password_resets", "id=$1", r.ID); err != nil {
			return err
		}
		send = func() {
			_ = hubmail.SendCard(context.Background(), s.A.Mail, e, "Your Hub verification code", hubmail.Card{Accent: hubmail.Blue, Title: "Your verification code", Preheader: "Your code is " + code,
				Lead: "Enter this code to continue. If you did not ask for it, you can ignore this message.", Rows: [][2]string{{"Your code is", code}, {"Works for", "10 minutes"}}, Note: "Never share this code with anyone. Hub staff will never ask for it."})
		}
		return nil
	})
	if err != nil {
		return uuid.Nil, err
	}
	if send != nil {
		go send() // the reply does not wait for the mail server
	}
	return id, nil
}

// ResetVerifyCode checks step 2 and returns the account's three security questions. An account whose questions were never set
// (made by an administrator or the setup tools) skips them: the link is emailed straight away, and `skipped` says so.
func (s *Service) ResetVerifyCode(ctx context.Context, id uuid.UUID, code string) (qs []string, skipped bool, err error) {
	var outErr error = errCode
	var send func()
	err = s.A.DB.Tx(ctx, func(tx pgx.Tx) error {
		x, err := scanReset(tx.QueryRow(ctx, `select `+resetCols+` from hub.password_resets where id=$1 for update`, id))
		if errors.Is(err, pgx.ErrNoRows) {
			s.codeMAC("otp", id, code)
			return nil
		}
		if err != nil {
			return err
		}
		if x.Stage != "otp" || time.Now().After(x.OTPExpires) {
			return nil
		}
		x.OTPTries++
		if !crypto.Equal(x.OTPHash, s.codeMAC("otp", id, strings.TrimSpace(code))) {
			if x.OTPTries >= 5 {
				x.Stage = "dead"
			}
			return s.saveReset(ctx, tx, x)
		}
		x.Stage = "quiz"
		if err := s.saveReset(ctx, tx, x); err != nil {
			return err
		}
		rs, err := tx.Query(ctx, `select question from hub.security_quiz where account_id=$1 order by position`, *x.Account)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var q string
			if err := rs.Scan(&q); err != nil {
				return err
			}
			qs = append(qs, q)
		}
		if err := rs.Err(); err != nil {
			return err
		}
		rs.Close()
		if quizNotSet(qs) {
			var e error
			if send, e = s.issueLink(ctx, tx, x); e != nil {
				return e
			}
			skipped, qs = true, nil
		}
		outErr = nil
		return nil
	})
	if err != nil {
		return nil, false, err
	}
	if send != nil {
		go send() // the reply does not wait for the mail server
	}
	return qs, skipped, outErr
}

// ResetQuiz checks step 3. On success the single-use link and its code are emailed.
func (s *Service) ResetQuiz(ctx context.Context, id uuid.UUID, answers []string) error {
	var outErr error = errQuiz
	var send func()
	err := s.A.DB.Tx(ctx, func(tx pgx.Tx) error {
		x, err := scanReset(tx.QueryRow(ctx, `select `+resetCols+` from hub.password_resets where id=$1 for update`, id))
		if errors.Is(err, pgx.ErrNoRows) {
			crypto.VerifyPassword("x", dummyHash)
			return nil
		}
		if err != nil {
			return err
		}
		if x.Stage != "quiz" || x.Account == nil {
			return nil
		}
		hs, err := tx.Query(ctx, `select answer_hash from hub.security_quiz where account_id=$1 order by position`, *x.Account)
		if err != nil {
			return err
		}
		var hashes []string
		for hs.Next() {
			var h string
			if err := hs.Scan(&h); err != nil {
				hs.Close()
				return err
			}
			hashes = append(hashes, h)
		}
		hs.Close()
		ok := len(answers) == 3 && len(hashes) == 3
		for i := 0; i < 3 && i < len(hashes); i++ { // check all three even after a miss: same time either way
			a := ""
			if i < len(answers) {
				a = normAnswer(answers[i])
			}
			if !crypto.VerifyPassword(a, hashes[i]) {
				ok = false
			}
		}
		if !ok {
			x.QuizTries++
			if x.QuizTries >= 3 {
				x.Stage = "dead"
			}
			return s.saveReset(ctx, tx, x)
		}
		var e error
		if send, e = s.issueLink(ctx, tx, x); e != nil {
			return e
		}
		outErr = nil
		return nil
	})
	if err != nil {
		return err
	}
	if send != nil {
		go send() // the reply does not wait for the mail server
	}
	return outErr
}

// ResetComplete performs step 5.
func (s *Service) ResetComplete(ctx context.Context, token, code, newPassword string) error {
	parts := strings.SplitN(token, ".", 2)
	id, err := uuid.Parse(parts[0])
	if err != nil || len(parts) != 2 {
		return errLink
	}
	if err := CheckPassword(newPassword); err != nil {
		return err
	}
	var outErr error = errLink
	var killed []uuid.UUID
	var notify string
	err = s.A.DB.Tx(ctx, func(tx pgx.Tx) error {
		x, err := scanReset(tx.QueryRow(ctx, `select `+resetCols+` from hub.password_resets where id=$1 for update`, id))
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		if x.Stage != "link" || x.LinkExpires == nil || time.Now().After(*x.LinkExpires) || x.Account == nil {
			return nil
		}
		x.LinkTries++
		okTok := crypto.Equal(x.LinkHash, hashSecret(parts[1]))
		okCode := crypto.Equal(x.LinkOTPHash, s.codeMAC("link", id, strings.TrimSpace(code)))
		if !okTok {
			x.Stage = "dead" // a wrong token is not a typo: kill the link
			return s.saveReset(ctx, tx, x)
		}
		if !okCode {
			if x.LinkTries >= 5 {
				x.Stage = "dead"
			}
			if err := s.saveReset(ctx, tx, x); err != nil {
				return err
			}
			outErr = errCode
			return nil
		}
		a, err := s.accountByID(ctx, tx, *x.Account, true)
		if err != nil {
			return err
		}
		a.PassHash, a.PasswordChangedAt, a.FailedAttempts, a.LockedUntil, a.MustChange = crypto.HashPassword(newPassword, s.A.PW), app.Now(), 0, nil, false
		if err := s.saveAccount(ctx, tx, a); err != nil {
			return err
		}
		x.Stage = "done"
		if err := s.saveReset(ctx, tx, x); err != nil {
			return err
		}
		if killed, err = s.revokeAllFor(ctx, tx, a.ID, uuid.Nil); err != nil {
			return err
		}
		notify, _ = s.decEmail(a)
		outErr = nil
		return nil
	})
	s.markRevoked(ctx, killed)
	if err != nil {
		return err
	}
	if notify != "" {
		_ = hubmail.SendCard(context.Background(), s.A.Mail, notify, "Your Hub password was changed", hubmail.Card{Accent: hubmail.Red, Title: "Your password was changed", Preheader: "You were signed out everywhere",
			Lead: "Your password was just changed and you were signed out on every device. If this was you, no action is needed.", Note: "If this was not you, reset your password again now and tell your administrator."})
	}
	return outErr
}

// issueLink moves a reset to its last stage: the single-use link and its code are made and saved, and the email carrying them is
// returned to be sent after the transaction commits.
func (s *Service) issueLink(ctx context.Context, tx pgx.Tx, x Reset) (func(), error) {
	id := x.ID
	secret, code := crypto.RandomToken(32), crypto.Digits(6)
	exp := app.Now().Add(30 * time.Minute)
	x.Stage, x.LinkHash, x.LinkOTPHash, x.LinkExpires = "link", hashSecret(secret), s.codeMAC("link", id, code), &exp
	if err := s.saveReset(ctx, tx, x); err != nil {
		return nil, err
	}
	a, err := s.accountByID(ctx, tx, *x.Account, false)
	if err != nil {
		return nil, err
	}
	email, err := s.decEmail(a)
	if err != nil {
		return nil, err
	}
	link := fmt.Sprintf("%s/reset-password?token=%s.%s", s.siteBase(), id, secret)
	return func() {
		_ = hubmail.SendCard(context.Background(), s.A.Mail, email, "Reset your Hub password", hubmail.Card{Accent: hubmail.Blue, Title: "Reset your password", Preheader: "Choose a new password",
			Lead: "We got a request to reset your password. Open the link and enter the code when asked.", Rows: [][2]string{{"Your code is", code}, {"Link works", "once, for 30 minutes"}}, Action: [2]string{"Choose a new password", link}, Note: "If you did not ask for this, ignore this email. Your password stays the same."})
	}, nil
}

// quizNotSet reports whether the questions are the placeholders given to accounts made by an administrator or the setup tools:
// nobody knows those answers, so the questions step cannot be asked of them. The emailed code and the emailed link still are.
func quizNotSet(qs []string) bool {
	if len(qs) == 0 {
		return true
	}
	for _, q := range qs {
		if !strings.HasPrefix(q, "Set in profile") {
			return false
		}
	}
	return true
}

package auth

import (
	"context"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/crypto"
	"hub/internal/httpx"
	hubmail "hub/internal/mail"
)

// One phone per account, like a chat app. The phone is known by the key it keeps in its secure storage (see device.go). The
// phone the account last confirmed stays signed in: no inactivity limit and a long, renewing sign-in, so it is never signed out
// on its own. Signing in on a different phone (a new key: another phone, or the app reinstalled) needs a code emailed to the
// account; once confirmed, that phone becomes the account's phone and the previous one is signed out and can do nothing until it
// goes through the same confirmation. A browser is separate: one browser and one phone may be signed in at the same time.

const (
	phoneCodeTTL    = 10 * time.Minute
	phoneCodeResend = time.Minute
	phoneCodeTries  = 5
)

var (
	errPhoneConfirm = httpx.Err(401, "device_confirm", "To confirm this phone, enter the 6-digit code we just emailed you.")
	errPhoneCode    = httpx.Err(401, "device_code", "That code is not right. Check the latest email from Hub and try again.")
	errPhoneExpired = httpx.Err(401, "device_code_expired", "That code has expired or was tried too often. Sign in again to get a new one.")
)

// phoneSignIn marks a sign-in bound to a phone in the context: it lives longer and has no inactivity limit.
type phoneSignIn struct{}

func asPhone(ctx context.Context) context.Context { return context.WithValue(ctx, phoneSignIn{}, true) }
func isPhone(ctx context.Context) bool            { v, _ := ctx.Value(phoneSignIn{}).(bool); return v }

// signInEnds is when a new session of this sign-in stops: the phone's is renewed on every use; a browser's keeps the end it got
// when it signed in (once a week the website asks for the password again), however often it is renewed.
type familyEnd struct{}

func keepEnd(ctx context.Context, end time.Time) context.Context {
	return context.WithValue(ctx, familyEnd{}, end)
}

func (s *Service) signInEnds(ctx context.Context) time.Time {
	if isPhone(ctx) {
		return time.Now().Add(s.A.Cfg.PhoneRefreshTTL)
	}
	if end, ok := ctx.Value(familyEnd{}).(time.Time); ok {
		return end
	}
	return time.Now().Add(s.A.Cfg.RefreshTTL)
}

type phoneCode struct {
	MAC   []byte    `json:"mac"`
	Tries int       `json:"tries"`
	At    time.Time `json:"at"`
}

func phoneCodeKey(account uuid.UUID, key []byte) string {
	return "phonecode:" + account.String() + ":" + hex.EncodeToString(hashSecret(string(key)))
}

// knownPhone reports whether this key is the account's phone. An account confirmed no phone yet keeps the phone it was already
// bound to before confirmation existed, so nobody signed in on the app has to confirm it again.
func knownPhone(ctx context.Context, tx pgx.Tx, account uuid.UUID, key []byte) (bool, error) {
	var current []byte
	err := tx.QueryRow(ctx, `select public_key from hub.account_phones where account_id=$1`, account).Scan(&current)
	if err == nil {
		return subtle.ConstantTimeCompare(current, key) == 1, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return false, err
	}
	var bound bool
	err = tx.QueryRow(ctx, `select exists (select 1 from hub.session_devices where account_id=$1 and public_key=$2)`, account, key).Scan(&bound)
	return bound, err
}

// confirmPhone lets a sign-in from the account's own phone through, and asks any other phone for the emailed code.
func (s *Service) confirmPhone(ctx context.Context, tx pgx.Tx, account uuid.UUID, email, name string, key []byte, code string) error {
	known, err := knownPhone(ctx, tx, account, key)
	if err != nil || known {
		return err
	}
	rk := phoneCodeKey(account, key)
	var pending phoneCode
	raw, gerr := s.A.Rdb.Get(ctx, rk).Bytes()
	have := gerr == nil && json.Unmarshal(raw, &pending) == nil
	if code == "" {
		if have && time.Since(pending.At) < phoneCodeResend {
			return errPhoneConfirm // a code went out moments ago: not another email
		}
		fresh := crypto.Digits(6)
		b, _ := json.Marshal(phoneCode{MAC: s.codeMAC("phone", account, fresh), At: time.Now()})
		if err := s.A.Rdb.Set(ctx, rk, b, phoneCodeTTL).Err(); err != nil {
			return err
		}
		go s.mailPhoneCode(email, name, fresh)
		return errPhoneConfirm
	}
	if !have {
		return errPhoneExpired
	}
	if subtle.ConstantTimeCompare(pending.MAC, s.codeMAC("phone", account, code)) != 1 {
		pending.Tries++
		if pending.Tries >= phoneCodeTries {
			s.A.Rdb.Del(ctx, rk)
			return errPhoneExpired
		}
		b, _ := json.Marshal(pending)
		s.A.Rdb.Set(ctx, rk, b, time.Until(pending.At.Add(phoneCodeTTL)))
		return errPhoneCode
	}
	s.A.Rdb.Del(ctx, rk)
	return nil
}

// registerPhone makes this key the account's phone (the previous phone's sign-in is ended by the caller).
func registerPhone(ctx context.Context, tx pgx.Tx, account uuid.UUID, key []byte, platform string) error {
	if len(platform) > 40 {
		platform = platform[:40]
	}
	_, err := tx.Exec(ctx, `insert into hub.account_phones (account_id, public_key, platform) values ($1,$2,$3)
		on conflict (account_id) do update set public_key=excluded.public_key, platform=excluded.platform, bound_at=now()`, account, key, platform)
	return err
}

func (s *Service) mailPhoneCode(email, name, code string) {
	_ = hubmail.SendCard(context.Background(), s.A.Mail, email, "Confirm your phone for Hub", hubmail.Card{Accent: hubmail.Blue, Title: "Confirm this phone",
		Preheader: "Your code is " + code, Greeting: "Hello " + name + ",",
		Lead: "Someone signed in to your Hub account in the Hub app on a phone it has not used before. Enter this code in the app to make it your phone. Your previous phone will be signed out.",
		Rows: [][2]string{{"Your code is", code}, {"Works for", "10 minutes"}},
		Note: "If this was not you, do not share the code: change your password now. Hub staff will never ask for this code."})
}

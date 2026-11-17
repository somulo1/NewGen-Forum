package auth

import (
	"context"
	"crypto/sha256"
	"errors"
	"net/mail"
	"strings"
	"time"
	"unicode"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pquerna/otp/totp"

	"hub/internal/app"
	"hub/internal/crypto"
	"hub/internal/httpx"
)

type Service struct{ A *app.App }

func New(a *app.App) *Service { return &Service{A: a} }

type QuizIn struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

type OrgRef struct {
	ID     uuid.UUID `json:"id"`
	Slug   string    `json:"slug"`
	Name   string    `json:"name"`
	Person uuid.UUID `json:"person,omitempty"`
}

type Tokens struct {
	Access     string    `json:"access_token"`
	Refresh    string    `json:"refresh_token"`
	ExpiresIn  int       `json:"expires_in"`
	Account    uuid.UUID `json:"account"`
	Name       string    `json:"name"`
	Org        *OrgRef   `json:"org,omitempty"`
	Orgs       []OrgRef  `json:"orgs"`
	MustChange bool      `json:"must_change_password,omitempty"`
}

// Claims are carried in the short-lived access token.
type Claims struct {
	jwt.RegisteredClaims
	Org  string `json:"org,omitempty"`
	Pid  string `json:"pid,omitempty"`
	Sid  string `json:"sid"`
	Name string `json:"name,omitempty"`
	MC   bool   `json:"mc,omitempty"` // must change password before anything else
	Ph   bool   `json:"ph,omitempty"` // a sign-in bound to the account's phone: no inactivity limit (see phone.go)
}

var (
	errCreds  = httpx.Err(401, "invalid_credentials", "Email or password is not right.")
	errLocked = httpx.Err(429, "account_locked", "Too many attempts. Try again later.")
)

// ---- validation ----

// domainAllowed applies HUB_ALLOWED_EMAIL_DOMAINS (any domain when it is not set).
func (s *Service) domainAllowed(email string) bool {
	if len(s.A.Cfg.EmailDomains) == 0 {
		return true
	}
	_, d, _ := strings.Cut(email, "@")
	for _, x := range s.A.Cfg.EmailDomains {
		if d == x || strings.HasSuffix(d, "."+x) {
			return true
		}
	}
	return false
}

var errDomain = httpx.Err(403, "domain_not_allowed", "This email domain cannot be used here. Use your work email.")

func NormEmail(e string) (string, error) {
	e = strings.ToLower(strings.TrimSpace(e))
	a, err := mail.ParseAddress(e)
	if err != nil || a.Address != e || len(e) > 254 || !strings.Contains(e[strings.LastIndex(e, "@"):], ".") {
		return "", httpx.Bad("Enter a valid email address.")
	}
	return e, nil
}

// CheckPassword enforces the password rules shown to people: 8+ characters with upper case, lower case and a number.
func CheckPassword(pw string) error {
	if len(pw) < 8 || len(pw) > 128 {
		return httpx.Bad("Use a password of 8 to 128 characters.")
	}
	var up, low, dig bool
	for _, r := range pw {
		switch {
		case unicode.IsUpper(r):
			up = true
		case unicode.IsLower(r):
			low = true
		case unicode.IsDigit(r):
			dig = true
		}
	}
	if !(up && low && dig) {
		return httpx.Bad("Your password needs an uppercase letter, a lowercase letter and a number.")
	}
	return nil
}

func normAnswer(a string) string { return strings.Join(strings.Fields(strings.ToLower(a)), " ") }

func (s *Service) emailIdx(email string) []byte {
	return crypto.Blind(s.A.Cfg.BlindKey, "email", email)
}

func (s *Service) encEmail(id uuid.UUID, email string) ([]byte, error) {
	return crypto.Field(s.A.PlatformKey(), crypto.Context{Org: "platform", Table: "accounts", Column: "email_enc", RowID: id.String()}, []byte(email))
}

func (s *Service) decEmail(a Account) (string, error) {
	b, err := crypto.FieldOpen(s.A.PlatformKey(), crypto.Context{Org: "platform", Table: "accounts", Column: "email_enc", RowID: a.ID.String()}, a.EmailEnc)
	return string(b), err
}

// ---- signup ----

func (s *Service) Signup(ctx context.Context, name, email, password string, quiz []QuizIn, ip string) (uuid.UUID, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 120 {
		return uuid.Nil, httpx.Bad("Enter your name.")
	}
	e, err := NormEmail(email)
	if err != nil {
		return uuid.Nil, err
	}
	if !s.domainAllowed(e) {
		return uuid.Nil, errDomain
	}
	if err := CheckPassword(password); err != nil {
		return uuid.Nil, err
	}
	if len(quiz) != 3 {
		return uuid.Nil, httpx.Bad("Choose three security questions.")
	}
	seen := map[string]bool{}
	for _, q := range quiz {
		if strings.TrimSpace(q.Question) == "" || len(normAnswer(q.Answer)) < 2 || seen[q.Question] {
			return uuid.Nil, httpx.Bad("Choose three different security questions and answer each one.")
		}
		seen[q.Question] = true
	}
	id := app.NewID()
	ct, err := s.encEmail(id, e)
	if err != nil {
		return uuid.Nil, err
	}
	acct := Account{ID: id, EmailIdx: s.emailIdx(e), EmailEnc: ct, DisplayName: name, PassHash: crypto.HashPassword(password, s.A.PW), Status: "active", PasswordChangedAt: app.Now()}
	err = s.A.DB.Tx(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `insert into hub.accounts (id, email_idx, email_enc, display_name, pass_hash, status, mfa_enabled, failed_attempts, password_changed_at) values ($1,$2,$3,$4,$5,'active',false,0,$6)`,
			acct.ID, acct.EmailIdx, acct.EmailEnc, acct.DisplayName, acct.PassHash, acct.PasswordChangedAt)
		if err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "accounts", "id=$1", id); err != nil {
			return err
		}
		for i, q := range quiz {
			row := Quiz{Account: id, Position: i + 1, Question: q.Question, AnswerHash: crypto.HashPassword(normAnswer(q.Answer), s.A.PW)}
			if _, err := tx.Exec(ctx, `insert into hub.security_quiz (account_id, position, question, answer_hash) values ($1,$2,$3,$4)`, id, row.Position, row.Question, row.AnswerHash); err != nil {
				return err
			}
		}
		return s.A.Sign.Seal(ctx, tx, "security_quiz", "account_id=$1", id)
	})
	if err != nil {
		if strings.Contains(err.Error(), "accounts_email_idx_key") {
			// do not reveal that the email is taken beyond what signup inherently must say
			return uuid.Nil, httpx.Err(409, "email_in_use", "An account with this email already exists. Try signing in.")
		}
		return uuid.Nil, err
	}
	return id, nil
}

// ---- login ----

type LoginIn struct {
	Email, Password, TOTP, Org, IP, UserAgent string
	Browser                                   string       // the browser's own random token (empty: a browser that has none yet)
	Device                                    *DeviceProof // from the phone app: the key this sign-in is bound to (see device.go)
	DeviceCode                                string       // the emailed code that confirms a phone the account has not used before
}

var dummyHash = crypto.HashPassword("dummy-password-for-timing", crypto.PasswordParams{Time: 3, MemoryKiB: 64 * 1024, Threads: 2})

func (s *Service) Login(ctx context.Context, in LoginIn) (*Tokens, error) {
	email, err := NormEmail(in.Email)
	if err != nil {
		return nil, errCreds
	}
	if !s.domainAllowed(email) {
		return nil, errDomain
	}
	devKey, phoneID, err := loginKey(in.Device, email)
	if err != nil {
		return nil, err
	}
	// brute-force guard: five wrong attempts (wrong password, or an email nobody has) from one address block that address for
	// two minutes; the block is recorded for the Security page. Each attempt is counted before its password is checked, with one
	// atomic increment, so a burst of parallel guesses cannot all slip past an empty block: only the first five are processed. A
	// sign-in that works gives its count back.
	var tried int64
	if in.IP != "" {
		var n int
		if err := s.A.DB.Tx(ctx, func(tx pgx.Tx) error {
			return tx.QueryRow(ctx, `select count(*) from hub.login_blocks where ip=$1 and until > now()`, in.IP).Scan(&n)
		}); err == nil && n > 0 {
			return nil, errLocked
		}
		var ok bool
		if tried, ok, _ = s.A.Limit.Take(ctx, "login-try", in.IP, maxTries, blockFor, true); !ok {
			s.recordBlock(ctx, "address", s.emailIdx(email), email, in.IP, "Too many sign-in attempts from this address", blockFor)
			return nil, errLocked
		}
	}
	var tok *Tokens
	var newAddr bool
	var loginErr error
	var replaced []uuid.UUID
	var ev struct {
		acct               uuid.UUID
		name, action, rslt string
		locked             *time.Time
	}
	err = s.A.DB.Tx(ctx, func(tx pgx.Tx) error {
		a, err := s.accountByIdx(ctx, tx, s.emailIdx(email), true)
		if errors.Is(err, pgx.ErrNoRows) {
			crypto.VerifyPassword(in.Password, dummyHash) // equalise timing so unknown emails cannot be told apart
			loginErr = errCreds
			return nil
		}
		if err != nil {
			return err
		}
		ev.acct, ev.name = a.ID, a.DisplayName
		if a.LockedUntil != nil && a.LockedUntil.After(time.Now()) {
			loginErr = errLocked
			ev.action, ev.rslt = "Sign-in blocked: account locked", "Blocked"
			// make sure this lock has a record for the Security page (one is kept while it lasts), even when it began
			// before blocks were recorded
			ev.locked = a.LockedUntil
			return nil
		}
		if !crypto.VerifyPassword(in.Password, a.PassHash) {
			ev.action, ev.rslt = "Failed sign-in (wrong password)", "Failed"
			a.FailedAttempts++
			if a.FailedAttempts >= 5 {
				t := app.Now().Add(blockFor)
				a.LockedUntil = &t
				a.FailedAttempts = 0
				ev.locked = &t
			}
			if err := s.saveAccount(ctx, tx, a); err != nil {
				return err
			}
			loginErr = errCreds
			return nil
		}
		if a.Status != "active" {
			loginErr = httpx.Err(403, "account_"+a.Status, "This account cannot sign in. Contact your administrator.")
			return nil
		}
		if a.MFAEnabled {
			secret, err := crypto.FieldOpen(s.A.PlatformKey(), crypto.Context{Org: "platform", Table: "accounts", Column: "mfa_secret_enc", RowID: a.ID.String()}, a.MFASecretEnc)
			if err != nil {
				return err
			}
			if in.TOTP == "" {
				loginErr = httpx.Err(401, "mfa_required", "Enter the code from your authenticator app.")
				return nil
			}
			if !totp.Validate(in.TOTP, string(secret)) {
				loginErr = errCreds
				return nil
			}
		}
		if a.FailedAttempts != 0 || a.LockedUntil != nil {
			a.FailedAttempts, a.LockedUntil = 0, nil
			if err := s.saveAccount(ctx, tx, a); err != nil {
				return err
			}
		}
		// a phone this account has not used before is confirmed with an emailed code first (see phone.go)
		if devKey != nil {
			if loginErr = s.confirmPhone(ctx, tx, a.ID, email, a.DisplayName, devKey, phoneID, in.DeviceCode); loginErr != nil {
				if loginErr == errPhoneConfirm {
					ev.action, ev.rslt = "Sign-in on a new phone: confirmation code emailed", "Pending"
				}
				return nil
			}
		}
		// a sign-in from a browser or device this account has never used is reported to its owner. A known browser on another network
		// or in another place is not: that is the same person moving about, and the email would only be noise.
		var browser []byte
		if len(in.Browser) > 200 {
			in.Browser = in.Browser[:200]
		}
		if in.Browser != "" {
			browser = hashSecret("browser:" + in.Browser)
		}
		{
			// until an account has signed in once from a browser that sends its token (just after this change), the old rule stands in:
			// an address it used before counts as known, so nobody gets one email per browser on the day this ships
			var known, any, tokened, seenIP bool
			if err := tx.QueryRow(ctx, `select $2::bytea is not null and exists (select 1 from hub.sessions where account_id=$1 and browser_hash=$2),
				exists (select 1 from hub.sessions where account_id=$1),
				exists (select 1 from hub.sessions where account_id=$1 and browser_hash is not null),
				$3 <> '' and exists (select 1 from hub.sessions where account_id=$1 and ip = nullif($3,'')::inet)`, a.ID, browser, in.IP).Scan(&known, &any, &tokened, &seenIP); err == nil {
				newAddr = any && !known && (tokened || !seenIP)
			}
		}
		ctx = context.WithValue(ctx, browserKey{}, browser)
		// one browser and one phone per account: signing in here signs out the previous one of the same kind
		if replaced, err = s.revokeSlot(ctx, tx, a.ID, devKey != nil); err != nil {
			return err
		}
		if devKey != nil {
			ctx = asPhone(ctx)
		}
		family := uuid.New()
		tok, loginErr = s.issue(ctx, tx, a, in.Org, family, in.IP, in.UserAgent)
		if loginErr == nil && devKey != nil {
			if err := bindFamily(ctx, tx, family, a.ID, devKey, in.Device.Platform); err != nil {
				return err
			}
			if err := s.registerPhone(ctx, tx, a.ID, devKey, phoneID, in.Device.Platform); err != nil {
				return err
			}
		}
		if loginErr == nil {
			ev.action, ev.rslt = "Signed in", "Success"
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if ev.action != "" {
		s.A.AuditAccount(ctx, ev.acct, ev.name, ev.action, ev.rslt, in.IP)
	}
	if ev.locked != nil {
		s.recordBlock(ctx, "account", s.emailIdx(email), email, in.IP, "Five wrong passwords in a row", time.Until(*ev.locked))
		go s.mailLocked(email, ev.name, in.IP, *ev.locked)
	}
	if loginErr == errCreds && tried >= maxTries {
		// the fifth attempt of the window failed: the address is blocked for a full two minutes from now (not from the window's start)
		kind, why := "address", "Five wrong sign-in attempts from this address"
		if ev.acct != uuid.Nil {
			why = "Five wrong sign-in attempts from this address on one account"
		}
		s.recordBlock(ctx, kind, s.emailIdx(email), email, in.IP, why, blockFor)
	}
	if loginErr == nil && in.IP != "" {
		s.A.Limit.Refund(ctx, "login-try", in.IP, blockFor)
	}
	if loginErr != nil {
		return nil, loginErr
	}
	if sid, err := uuid.Parse(strings.SplitN(tok.Refresh, ".", 2)[0]); err == nil {
		why := whyReplaced
		if devKey != nil {
			why = whyPhone
		}
		s.markEnded(ctx, replaced, why)
		s.touch(ctx, sid)
		if tok.Org != nil {
			s.SeenToday(ctx, tok.Org.ID, tok.Org.Person)
		}
		if len(replaced) > 0 {
			s.tellReplaced(ctx, ev.acct, sid)
		}
		if newAddr {
			go s.mailNewAddress(email, ev.name, in.IP, in.UserAgent)
		}
	}
	return tok, nil
}

// orgsFor lists the organizations an account belongs to and picks the active one.
func (s *Service) orgsFor(ctx context.Context, tx pgx.Tx, acct uuid.UUID, want string) ([]OrgRef, *OrgRef, error) {
	rs, err := tx.Query(ctx, `select o.id, o.slug, o.name from hub.account_orgs ao join hub.orgs o on o.id=ao.org_id where ao.account_id=$1 and o.status='active' order by o.created_at`, acct)
	if err != nil {
		return nil, nil, err
	}
	var list []OrgRef
	for rs.Next() {
		var o OrgRef
		if err := rs.Scan(&o.ID, &o.Slug, &o.Name); err != nil {
			rs.Close()
			return nil, nil, err
		}
		list = append(list, o)
	}
	rs.Close()
	var active *OrgRef
	for i := range list {
		if want == "" || list[i].Slug == want {
			active = &list[i]
			break
		}
	}
	if want != "" && active == nil {
		return list, nil, httpx.Err(403, "not_a_member", "You do not belong to that organization.")
	}
	return list, active, nil
}

func (s *Service) issue(ctx context.Context, tx pgx.Tx, a Account, orgSlug string, family uuid.UUID, ip, ua string) (*Tokens, error) {
	orgs, active, err := s.orgsFor(ctx, tx, a.ID, orgSlug)
	if err != nil {
		return nil, err
	}
	var person uuid.UUID
	var orgID *uuid.UUID
	if active != nil {
		// the person record for this organization decides whether the account may enter it
		var status string
		err := tx.QueryRow(ctx, `select set_config('hub.org_id', $1, true)`, active.ID.String()).Scan(new(string))
		if err != nil {
			return nil, err
		}
		if err := tx.QueryRow(ctx, `select id, status from hub.people where account_id=$1`, a.ID).Scan(&person, &status); err != nil {
			return nil, httpx.Err(403, "no_profile", "You have no profile in this organization yet.")
		}
		if status != "active" {
			return nil, httpx.Err(403, "account_"+status, "Your access to this organization is turned off. Contact your administrator.")
		}
		active.Person = person
		orgID = &active.ID
	}
	sid := app.NewID()
	secret := crypto.RandomToken(32)
	rt := Session{ID: sid, Account: a.ID, Family: family, Org: orgID, RefreshHash: hashSecret(secret), ExpiresAt: s.signInEnds(ctx).UTC()}
	rt.Browser, _ = ctx.Value(browserKey{}).([]byte)
	if err := s.insertSession(ctx, tx, rt, ip, ua); err != nil {
		return nil, err
	}
	// an access token never outlives its sign-in: a browser is signed out on the second its week is up
	exp := time.Now().Add(s.A.Cfg.AccessTTL)
	if rt.ExpiresAt.Before(exp) {
		exp = rt.ExpiresAt
	}
	access, err := s.sign(a, active, sid, isPhone(ctx), exp)
	if err != nil {
		return nil, err
	}
	return &Tokens{Access: access, Refresh: sid.String() + "." + secret, ExpiresIn: int(time.Until(exp).Seconds()), Account: a.ID, Name: a.DisplayName, Org: active, Orgs: orgs, MustChange: a.MustChange}, nil
}

type browserKey struct{}

func hashSecret(secret string) []byte { h := sha256.Sum256([]byte(secret)); return h[:] }

func (s *Service) sign(a Account, org *OrgRef, sid uuid.UUID, phone bool, exp time.Time) (string, error) {
	now := time.Now()
	c := Claims{RegisteredClaims: jwt.RegisteredClaims{Issuer: "hub", Subject: a.ID.String(), IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(exp), ID: uuid.NewString()}, Sid: sid.String(), Name: a.DisplayName, MC: a.MustChange, Ph: phone}
	if org != nil {
		c.Org, c.Pid = org.ID.String(), org.Person.String()
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(s.A.Cfg.TokenKey)
}

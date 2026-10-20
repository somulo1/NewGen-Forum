package auth

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
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
	hubmail "hub/internal/mail"
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
	// brute-force guard: five wrong attempts (wrong password, or an email nobody has) from one address block that address for
	// two minutes; the block is recorded for the Security page. Counting happens below, on each failure.
	if in.IP != "" {
		var n int
		if err := s.A.DB.Tx(ctx, func(tx pgx.Tx) error {
			return tx.QueryRow(ctx, `select count(*) from hub.login_blocks where ip=$1 and until > now()`, in.IP).Scan(&n)
		}); err == nil && n > 0 {
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
		// a sign-in from an address this account has not used before is reported to its owner
		if in.IP != "" {
			var seen, any bool
			if err := tx.QueryRow(ctx, `select exists (select 1 from hub.sessions where account_id=$1 and ip=$2::inet), exists (select 1 from hub.sessions where account_id=$1)`, a.ID, in.IP).Scan(&seen, &any); err == nil {
				newAddr = any && !seen
			}
		}
		// one session per account: signing in here signs out every other browser and device
		if replaced, err = s.revokeAllFor(ctx, tx, a.ID, uuid.Nil); err != nil {
			return err
		}
		tok, loginErr = s.issue(ctx, tx, a, in.Org, uuid.New(), in.IP, in.UserAgent)
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
	if loginErr == errCreds {
		// the fifth failure within two minutes (this one included) blocks the address
		if ok, retry := s.A.Limit.Allow(ctx, "login-fail", in.IP, 4, blockFor, true); !ok {
			kind, why := "address", "Five wrong sign-in attempts from this address"
			if ev.acct != uuid.Nil {
				why = "Five wrong sign-in attempts from this address on one account"
			}
			s.recordBlock(ctx, kind, s.emailIdx(email), email, in.IP, why, retry)
		}
	}
	if loginErr != nil {
		return nil, loginErr
	}
	if sid, err := uuid.Parse(strings.SplitN(tok.Refresh, ".", 2)[0]); err == nil {
		s.markEnded(ctx, replaced, whyReplaced)
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
	rt := Session{ID: sid, Account: a.ID, Family: family, Org: orgID, RefreshHash: hashSecret(secret), ExpiresAt: app.Now().Add(s.A.Cfg.RefreshTTL)}
	if err := s.insertSession(ctx, tx, rt, ip, ua); err != nil {
		return nil, err
	}
	access, err := s.sign(a, active, sid)
	if err != nil {
		return nil, err
	}
	return &Tokens{Access: access, Refresh: sid.String() + "." + secret, ExpiresIn: int(s.A.Cfg.AccessTTL.Seconds()), Account: a.ID, Name: a.DisplayName, Org: active, Orgs: orgs, MustChange: a.MustChange}, nil
}

func hashSecret(secret string) []byte { h := sha256.Sum256([]byte(secret)); return h[:] }

func (s *Service) sign(a Account, org *OrgRef, sid uuid.UUID) (string, error) {
	now := time.Now()
	c := Claims{RegisteredClaims: jwt.RegisteredClaims{Issuer: "hub", Subject: a.ID.String(), IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(s.A.Cfg.AccessTTL)), ID: uuid.NewString()}, Sid: sid.String(), Name: a.DisplayName, MC: a.MustChange}
	if org != nil {
		c.Org, c.Pid = org.ID.String(), org.Person.String()
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(s.A.Cfg.TokenKey)
}

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

// mailNewAddress tells the owner of an account that someone signed in from an address it had not used before.
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
	_ = hubmail.SendCard(context.Background(), s.A.Mail, email, "New sign-in to your Hub account", hubmail.Card{Greeting: "Hello " + firstWord(name) + ",", Title: "New sign-in to your account", Preheader: "A sign-in from " + ip,
		Lead: "Your account was just signed in to from an address you have not used before. Signing in here signed you out of your other browsers and devices.", Rows: rows,
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

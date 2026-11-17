package auth

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/httpx"
	"hub/internal/orgsettings"
	"hub/internal/realtime"
	"hub/internal/tenant"
)

// One browser and one phone per account, and sign-out after inactivity in the browser.
//
// Signing in on a browser ends the account's other browser session, and signing in on a phone ends its other phone session
// (like WhatsApp and WhatsApp Web); the one replaced is told over the realtime socket. The phone has no inactivity limit
// (see phone.go). Inactivity is measured from real use: the app reports the
// person's last click or key press (POST /auth/active); background requests do not count. When the organization's limit
// passes without activity, the next request is refused and the session ends, to the second.

const (
	whyReplaced    = "replaced"
	whyPhone       = "phone" // the account was confirmed on another phone
	whyBlocked     = "blocked"
	whyDeactivated = "deactivated"
	whyRemoved     = "removed"
	whyIdle        = "idle"
)

var (
	errReplaced = httpx.Err(401, "session_replaced", "You signed in on another browser or device, so you were signed out here.")
	errIdle     = httpx.Err(401, "session_expired", "You were signed out after being inactive. Sign in again to continue.")
	errWeek     = httpx.Err(401, "session_week", "For your security the website signs you out once a week. Sign in again to continue.")
	errBlocked  = httpx.Err(401, "access_blocked", "An administrator blocked your access to this organization, so you were signed out.")
	errOff      = httpx.Err(401, "access_deactivated", "An administrator turned off your access to this organization, so you were signed out.")
	errRemoved  = httpx.Err(401, "access_removed", "You were removed from this organization, so you were signed out.")
	errPhoneOut = httpx.Err(401, "phone_replaced", "Your account is now on another phone, so this phone was signed out. Signing in here again needs a code from your email.")
)

// markEnded revokes access tokens now and remembers why each session ended, so the browser can say so.
func (s *Service) markEnded(ctx context.Context, ids []uuid.UUID, why string) {
	for _, id := range ids {
		s.A.Rdb.Set(ctx, "revoked:"+id.String(), why, s.A.Cfg.AccessTTL+time.Minute)
		s.A.Rdb.Set(ctx, "ended:"+id.String(), why, 30*24*time.Hour) // a phone that was offline for weeks still learns why
	}
}

func endedErr(why string) error {
	switch why {
	case whyReplaced:
		return errReplaced
	case whyIdle:
		return errIdle
	case whyPhone:
		return errPhoneOut
	case whyBlocked:
		return errBlocked
	case whyDeactivated:
		return errOff
	case whyRemoved:
		return errRemoved
	}
	return httpx.ErrUnauthorized
}

// whyEnded says why a session ended, when that is known.
func (s *Service) whyEnded(ctx context.Context, sid uuid.UUID) error {
	why, _ := s.A.Rdb.Get(ctx, "ended:"+sid.String()).Result()
	return endedErr(why)
}

// IdleLimit is how long a person in this organization may be inactive before being signed out (0: no limit).
func (s *Service) IdleLimit(ctx context.Context, org uuid.UUID) time.Duration {
	if v, ok := orgsettings.IdleCache.Load(org); ok {
		if e := v.(orgsettings.IdleEntry); time.Since(e.At) < time.Minute {
			return e.D
		}
	}
	d := s.idleLimitFresh(ctx, org)
	orgsettings.IdleCache.Store(org, orgsettings.IdleEntry{D: d, At: time.Now()})
	return d
}

func (s *Service) idleLimitFresh(ctx context.Context, org uuid.UUID) time.Duration {
	key := "idle:" + org.String()
	if v, err := s.A.Rdb.Get(ctx, key).Result(); err == nil {
		n, _ := strconv.Atoi(v)
		return time.Duration(n) * time.Minute
	}
	sec := orgsettings.DefaultSecurity()
	_ = s.A.DB.WithOrg(ctx, org, func(tx pgx.Tx) error { return orgsettings.Get(ctx, s.A, tx, org, "security", &sec) })
	m := orgsettings.SessionMinutes(sec.Session)
	s.A.Rdb.Set(ctx, key, strconv.Itoa(m), 10*time.Minute)
	return time.Duration(m) * time.Minute
}

func actKey(sid uuid.UUID) string { return "act:" + sid.String() }

// touch records activity now for a session.
func (s *Service) touch(ctx context.Context, sid uuid.UUID) {
	s.A.Rdb.Set(ctx, actKey(sid), strconv.FormatInt(time.Now().UnixMilli(), 10), 25*time.Hour)
}

// lastActive returns when the session was last used by its person (zero when never recorded).
func (s *Service) lastActive(ctx context.Context, sid uuid.UUID) time.Time {
	v, err := s.A.Rdb.Get(ctx, actKey(sid)).Result()
	if err != nil {
		return time.Time{}
	}
	ms, _ := strconv.ParseInt(v, 10, 64)
	return time.UnixMilli(ms)
}

// idleExpired reports whether the session passed the organization's inactivity limit.
func (s *Service) idleExpired(ctx context.Context, org, sid uuid.UUID) bool {
	limit := s.IdleLimit(ctx, org)
	if limit <= 0 {
		return false
	}
	last := s.lastActive(ctx, sid)
	return last.IsZero() || time.Since(last) > limit
}

// endIdle ends a session that passed the inactivity limit.
func (s *Service) endIdle(ctx context.Context, sid uuid.UUID) {
	var ids []uuid.UUID
	_ = s.A.DB.Tx(ctx, func(tx pgx.Tx) error {
		x, err := scanSession(tx.QueryRow(ctx, `select `+sessionCols+` from hub.sessions where id=$1 for update`, sid))
		if err != nil || x.RevokedAt != nil {
			return nil
		}
		ids = append(ids, x.ID)
		return s.revokeSession(ctx, tx, x)
	})
	s.markEnded(ctx, append(ids, sid), whyIdle)
}

// limitFor is the inactivity limit of the caller's session: the organization's for a browser, none for the phone.
func (s *Service) limitFor(ctx context.Context, p *tenant.Principal) time.Duration {
	if p.Phone {
		return 0
	}
	return s.IdleLimit(ctx, p.Org)
}

// Active records that the person is using the app (a click or key press) and counts them as active today.
func (s *Service) Active(ctx context.Context, sid, org, person uuid.UUID, phone bool) error {
	if org != uuid.Nil && !phone && s.idleExpired(ctx, org, sid) {
		s.endIdle(ctx, sid)
		return errIdle
	}
	s.touch(ctx, sid)
	if org != uuid.Nil && person != uuid.Nil {
		s.SeenToday(ctx, org, person)
	}
	return nil
}

// SeenToday records that a person used the system today (once a day per person, however often they sign in).
func (s *Service) SeenToday(ctx context.Context, org, person uuid.UUID) {
	day := time.Now().In(nairobi).Format("2006-01-02")
	if ok, err := s.A.Rdb.SetNX(ctx, "seen:"+org.String()+":"+person.String()+":"+day, 1, 48*time.Hour).Result(); err == nil && !ok {
		return
	}
	added := false
	_ = s.A.DB.WithOrg(ctx, org, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `insert into hub.person_days (org_id, person_id, day) values ($1,$2,$3::date) on conflict do nothing`, org, person, day)
		added = err == nil && tag.RowsAffected() > 0
		return err
	})
	if added {
		// usage figures (Security & logs) move: told over the live connection, not found by polling
		s.A.Publish(ctx, realtime.Event{Type: "activity.changed", Channel: realtime.OrgChannel(org), Org: org.String()})
	}
}

var nairobi = func() *time.Location {
	l, err := time.LoadLocation("Africa/Nairobi")
	if err != nil {
		return time.UTC
	}
	return l
}()

// tellReplaced tells the account's other browsers that they were signed out by a new sign-in.
func (s *Service) tellReplaced(ctx context.Context, account, keep uuid.UUID) {
	d, _ := json.Marshal(map[string]any{"keep": keep})
	s.A.Publish(ctx, realtime.Event{Type: "session.replaced", Channel: realtime.AcctChannel(account), Data: d})
}

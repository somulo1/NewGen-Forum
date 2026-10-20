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
)

// One live session per account, and sign-out after inactivity.
//
// Signing in ends every other session of the account (another browser or device is signed out at once, like WhatsApp
// Web), and that browser is told over the realtime socket. Inactivity is measured from real use: the app reports the
// person's last click or key press (POST /auth/active); background requests do not count. When the organization's limit
// passes without activity, the next request is refused and the session ends, to the second.

const (
	whyReplaced = "replaced"
	whyIdle     = "idle"
)

var (
	errReplaced = httpx.Err(401, "session_replaced", "You signed in on another browser or device, so you were signed out here.")
	errIdle     = httpx.Err(401, "session_expired", "You were signed out after being inactive. Sign in again to continue.")
)

// markEnded revokes access tokens now and remembers why each session ended, so the browser can say so.
func (s *Service) markEnded(ctx context.Context, ids []uuid.UUID, why string) {
	for _, id := range ids {
		s.A.Rdb.Set(ctx, "revoked:"+id.String(), why, s.A.Cfg.AccessTTL+time.Minute)
		s.A.Rdb.Set(ctx, "ended:"+id.String(), why, 7*24*time.Hour)
	}
}

func endedErr(why string) error {
	switch why {
	case whyReplaced:
		return errReplaced
	case whyIdle:
		return errIdle
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

// Active records that the person is using the app (a click or key press) and counts them as active today.
func (s *Service) Active(ctx context.Context, sid, org, person uuid.UUID) error {
	if org != uuid.Nil && s.idleExpired(ctx, org, sid) {
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
	_ = s.A.DB.WithOrg(ctx, org, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `insert into hub.person_days (org_id, person_id, day) values ($1,$2,$3::date) on conflict do nothing`, org, person, day)
		return err
	})
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

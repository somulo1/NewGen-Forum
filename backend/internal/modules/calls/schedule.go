package calls

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/httpx"
	"hub/internal/tenant"
)

// A scheduled meeting or standup only has a room while it is on: people are let in from 5 minutes before the start until the end
// time, on the days it takes place (once, daily, on weekdays or weekly). Outside that, nothing is created, no pass is minted and no
// media room is reserved, so an early arrival costs the system nothing. Calls that belong to no calendar entry (a forum call, an
// ad-hoc meeting) are not scheduled and are never refused here.
const joinLead = 5 * time.Minute

// slot is one calendar entry's plain schedule (titles and notes stay sealed; the date and times are not).
type slot struct {
	Type, Date, Start, End, Repeat string
}

// entryOf finds the meeting or standup a conversation was made for, if it is one.
func entryOf(ctx context.Context, tx pgx.Tx, conv uuid.UUID) (*slot, error) {
	var raw []byte
	err := tx.QueryRow(ctx, `select meta from hub.records where container_id=$1 and collection='entry' and deleted_at is null
		and meta->>'type' in ('Meeting','Standup') order by created_at limit 1`, conv).Scan(&raw)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var s slot
	if json.Unmarshal(raw, &s) != nil || s.Start == "" {
		return nil, nil
	}
	return &s, nil
}

func clock(day time.Time, hm string, loc *time.Location) (time.Time, bool) {
	t, err := time.ParseInLocation("15:04", hm, loc)
	if err != nil {
		return time.Time{}, false
	}
	return time.Date(day.Year(), day.Month(), day.Day(), t.Hour(), t.Minute(), 0, 0, loc), true
}

// occursOn says whether the entry takes place on the calendar day of t.
func (s *slot) occursOn(t time.Time) bool {
	day := t.Format("2006-01-02")
	switch s.Repeat {
	case "", "None":
		return s.Date == day
	}
	if day < s.Date {
		return false
	}
	switch s.Repeat {
	case "Daily":
		return true
	case "Weekdays":
		return t.Weekday() != time.Saturday && t.Weekday() != time.Sunday
	case "Weekly":
		d, err := time.ParseInLocation("2006-01-02", s.Date, t.Location())
		return err == nil && d.Weekday() == t.Weekday()
	}
	return false
}

// window gives, for the day of now, when the room opens and when the entry ends.
func (s *slot) window(now time.Time) (opens, starts, ends time.Time, ok bool) {
	starts, ok = clock(now, s.Start, now.Location())
	if !ok {
		return
	}
	ends, ok = clock(now, s.End, now.Location())
	if !ok || !ends.After(starts) {
		ends = starts.Add(time.Hour) // no usable end time: an hour
	}
	return starts.Add(-joinLead), starts, ends, true
}

// open reports whether now is inside the room's open time.
func (s *slot) open(now time.Time) (bool, string) {
	if !s.occursOn(now) {
		return false, "This meeting is not on today."
	}
	opens, starts, ends, ok := s.window(now)
	if !ok {
		return true, ""
	}
	switch {
	case now.Before(opens):
		return false, fmt.Sprintf("This meeting opens at %s, 5 minutes before it starts at %s.", opens.Format("15:04"), starts.Format("15:04"))
	case !now.Before(ends):
		return false, fmt.Sprintf("This meeting ended at %s.", ends.Format("15:04"))
	}
	return true, ""
}

// checkOpen refuses to start or join a meeting or standup room outside its scheduled time.
func (m *Module) checkOpen(ctx context.Context, tx pgx.Tx, callContext string, conv *uuid.UUID, now time.Time) error {
	if conv == nil || (callContext != "meeting" && callContext != "standup") {
		return nil
	}
	s, err := entryOf(ctx, tx, *conv)
	if err != nil || s == nil {
		return err
	}
	loc, lerr := time.LoadLocation(m.A.Cfg.ScheduleTZ)
	if lerr != nil {
		loc = time.UTC
	}
	if ok, why := s.open(now.In(loc)); !ok {
		return httpx.Err(409, "not_open", why)
	}
	return nil
}

// runningRoom finds the room of this meeting that someone is in right now, so the second person to press "Join" enters the same
// room instead of opening another one.
func (m *Module) runningRoom(ctx context.Context, p *tenant.Principal, conv *uuid.UUID, callContext string) (uuid.UUID, bool) {
	if conv == nil || (callContext != "meeting" && callContext != "standup") {
		return uuid.Nil, false
	}
	var id uuid.UUID
	err := m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `select c.id from hub.calls c where c.conversation_id=$1 and c.context in ('meeting','standup') and c.status='active' and c.ended_at is null
			and c.started_at > now() - interval '12 hours'
			and exists (select 1 from hub.call_participants cp where cp.org_id = c.org_id and cp.call_id = c.id and cp.joined_at is not null and cp.left_at is null)
			order by c.started_at desc limit 1`, *conv).Scan(&id)
	})
	return id, err == nil
}

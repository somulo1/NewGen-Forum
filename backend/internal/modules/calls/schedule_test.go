package calls

import (
	"testing"
	"time"
)

func TestMeetingRoomOpensFiveMinutesBeforeUntilTheEnd(t *testing.T) {
	loc, _ := time.LoadLocation("Africa/Nairobi")
	at := func(day, hm string) time.Time {
		v, _ := time.ParseInLocation("2006-01-02 15:04", day+" "+hm, loc)
		return v
	}
	once := &slot{Type: "Meeting", Date: "2026-11-03", Start: "09:00", End: "10:00", Repeat: "None"}
	daily := &slot{Type: "Standup", Date: "2026-11-02", Start: "09:30", End: "09:45", Repeat: "Daily"}
	weekdays := &slot{Type: "Standup", Date: "2026-11-02", Start: "09:30", End: "09:45", Repeat: "Weekdays"}
	weekly := &slot{Type: "Meeting", Date: "2026-11-02", Start: "14:00", End: "15:00", Repeat: "Weekly"} // Mondays
	noEnd := &slot{Type: "Meeting", Date: "2026-11-03", Start: "09:00", Repeat: "None"}
	for _, c := range []struct {
		name string
		s    *slot
		now  time.Time
		want bool
	}{
		{"too early", once, at("2026-11-03", "08:54"), false},
		{"opens 5 minutes before", once, at("2026-11-03", "08:55"), true},
		{"during", once, at("2026-11-03", "09:30"), true},
		{"last minute", once, at("2026-11-03", "09:59"), true},
		{"at the end", once, at("2026-11-03", "10:00"), false},
		{"another day", once, at("2026-11-04", "09:30"), false},
		{"before the first day", daily, at("2026-11-01", "09:30"), false},
		{"daily, later day", daily, at("2026-12-10", "09:40"), true},
		{"daily, after end", daily, at("2026-12-10", "09:46"), false},
		{"weekdays, Saturday", weekdays, at("2026-11-07", "09:35"), false},
		{"weekdays, Friday", weekdays, at("2026-11-06", "09:35"), true},
		{"weekly, Monday", weekly, at("2026-11-09", "14:30"), true},
		{"weekly, Tuesday", weekly, at("2026-11-10", "14:30"), false},
		{"no end time lasts an hour", noEnd, at("2026-11-03", "09:59"), true},
		{"no end time, after the hour", noEnd, at("2026-11-03", "10:01"), false},
	} {
		if got, why := c.s.open(c.now); got != c.want {
			t.Errorf("%s: open=%v (%s), want %v", c.name, got, why, c.want)
		}
	}
}

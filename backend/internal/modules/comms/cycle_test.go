package comms

import (
	"testing"
	"time"
)

func TestAwardBoardResetsEverySaturdayMidnight(t *testing.T) {
	loc, _ := time.LoadLocation("Africa/Nairobi")
	at := func(y int, m time.Month, d, h, mi int) time.Time { return time.Date(y, m, d, h, mi, 0, 0, loc) }
	cases := []struct{ now, start time.Time }{
		{at(2026, 10, 19, 10, 0), at(2026, 10, 17, 0, 0)},  // Monday: cycle began Saturday the 17th
		{at(2026, 10, 24, 0, 0), at(2026, 10, 24, 0, 0)},   // Saturday 00:00 sharp: a new cycle starts
		{at(2026, 10, 23, 23, 59), at(2026, 10, 17, 0, 0)}, // a minute earlier: still the old one
		{at(2026, 10, 24, 15, 0), at(2026, 10, 24, 0, 0)},  // Saturday afternoon
		{at(2026, 10, 25, 9, 0), at(2026, 10, 24, 0, 0)},   // Sunday
	}
	for _, c := range cases {
		got := cycleAt(c.now, loc)
		if !got.Start.Equal(c.start) || !got.Next.Equal(c.start.AddDate(0, 0, 7)) {
			t.Fatalf("at %v: start %v next %v, want start %v", c.now, got.Start.In(loc), got.Next.In(loc), c.start)
		}
	}
}

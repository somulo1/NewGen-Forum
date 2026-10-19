package tickets

import (
	"strings"
	"time"
)

// Config is an organization's ticket timing policy.
type Config struct {
	Start    string `json:"start"` // working day start, "08:00"
	End      string `json:"end"`
	Timezone string `json:"timezone"`
	Pause    struct {
		Weekends bool `json:"weekends"`
		Holidays bool `json:"holidays"`
		Offhours bool `json:"offhours"`
		Onhold   bool `json:"onhold"`
		Awaiting bool `json:"awaiting"`
	} `json:"pause"`
	Notify struct {
		Solved   bool `json:"solved"`
		Overdue  bool `json:"overdue"`
		Nearing  bool `json:"nearing"`
		Assigned bool `json:"assigned"`
	} `json:"notify"`
	NearPct  int      `json:"nearPct"`
	Escalate string   `json:"escalate"`
	Holidays []string `json:"holidays"`
}

func defaultConfig() Config {
	c := Config{Start: "08:00", End: "17:00", Timezone: "Africa/Nairobi", NearPct: 80, Escalate: "Head of IT Operations"}
	c.Pause.Weekends, c.Pause.Holidays, c.Pause.Offhours, c.Pause.Onhold = true, true, true, true
	c.Notify.Solved, c.Notify.Overdue, c.Notify.Nearing, c.Notify.Assigned = true, true, true, true
	return c
}

func (c Config) loc() *time.Location {
	l, err := time.LoadLocation(c.Timezone)
	if err != nil {
		return time.UTC
	}
	return l
}

func hm(s string, def int) int {
	t, err := time.Parse("15:04", s)
	if err != nil {
		return def
	}
	return t.Hour()*60 + t.Minute()
}

// counts reports whether the clock is running at this local instant, under the policy.
func (c Config) counts(t time.Time) bool {
	if c.Pause.Weekends && (t.Weekday() == time.Saturday || t.Weekday() == time.Sunday) {
		return false
	}
	if c.Pause.Holidays {
		d := t.Format("2006-01-02")
		for _, h := range c.Holidays {
			if strings.TrimSpace(h) == d {
				return false
			}
		}
	}
	if c.Pause.Offhours {
		m := t.Hour()*60 + t.Minute()
		if m < hm(c.Start, 8*60) || m >= hm(c.End, 17*60) {
			return false
		}
	}
	return true
}

// Due adds `hours` of running clock to `from`. Minute resolution is plenty for a TAT measured in hours or days.
func (c Config) Due(from time.Time, hours int) time.Time {
	loc := c.loc()
	t := from.In(loc).Truncate(time.Minute)
	left := hours * 60
	for i := 0; left > 0 && i < 60*24*400; i++ { // never loops forever on a policy that never counts
		if c.counts(t) {
			left--
		}
		t = t.Add(time.Minute)
	}
	return t
}

// Running is how many minutes of the policy's clock lie between a and b.
func (c Config) Running(a, b time.Time) time.Duration {
	if !b.After(a) {
		return 0
	}
	loc := c.loc()
	t, end := a.In(loc).Truncate(time.Minute), b.In(loc)
	n := 0
	for ; t.Before(end) && n < 60*24*400; t = t.Add(time.Minute) {
		if c.counts(t) {
			n++
		}
	}
	return time.Duration(n) * time.Minute
}

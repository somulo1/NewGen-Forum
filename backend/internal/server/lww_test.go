package server

import (
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestWrittenAtAndTheFortyEightHourLimit(t *testing.T) {
	at := func(v string) (time.Time, bool) {
		r := httptest.NewRequest("PUT", "/v1/x", nil)
		if v != "" {
			r.Header.Set("X-Hub-Written-At", v)
		}
		return writtenAt(r)
	}
	ms := func(d time.Duration) string { return strconv.FormatInt(time.Now().Add(d).UnixMilli(), 10) }
	near := func(a, b time.Time) bool { d := a.Sub(b); return d < time.Second && d > -time.Second }

	if got, old := at(ms(-time.Hour)); !near(got, time.Now().Add(-time.Hour)) || old {
		t.Fatalf("an edit made offline an hour ago: %v expired=%v", got, old)
	}
	if _, old := at(ms(-47 * time.Hour)); old {
		t.Fatal("47 hours offline must still be applied")
	}
	if _, old := at(ms(-49 * time.Hour)); !old {
		t.Fatal("49 hours offline must be refused")
	}
	for _, v := range []string{"", "nonsense", ms(time.Hour)} {
		if got, old := at(v); !near(got, time.Now()) || old {
			t.Fatalf("%q should count as now, got %v expired=%v", v, got, old)
		}
	}
}

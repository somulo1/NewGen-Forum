package server

import (
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestWrittenAtTrustsOnlyBelievableTimes(t *testing.T) {
	at := func(v string) time.Time {
		r := httptest.NewRequest("PUT", "/v1/x", nil)
		if v != "" {
			r.Header.Set("X-Hub-Written-At", v)
		}
		return writtenAt(r)
	}
	near := func(a, b time.Time) bool { d := a.Sub(b); return d < time.Second && d > -time.Second }
	hourAgo := time.Now().Add(-time.Hour)
	if got := at(strconv.FormatInt(hourAgo.UnixMilli(), 10)); !near(got, hourAgo) {
		t.Fatalf("an edit made offline an hour ago counts as %v", got)
	}
	for _, v := range []string{"", "nonsense", strconv.FormatInt(time.Now().Add(time.Hour).UnixMilli(), 10), strconv.FormatInt(time.Now().Add(-8*24*time.Hour).UnixMilli(), 10)} {
		if got := at(v); !near(got, time.Now()) {
			t.Fatalf("%q should count as now, got %v", v, got)
		}
	}
}

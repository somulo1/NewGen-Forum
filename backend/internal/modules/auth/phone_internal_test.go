package auth

import (
	"context"
	"testing"
	"time"

	"hub/internal/app"
	"hub/internal/config"
)

// A browser's sign-in ends a week after signing in, however often it is renewed; the phone's is renewed on every use.
func TestSignInEnds(t *testing.T) {
	s := &Service{A: &app.App{Cfg: &config.Config{RefreshTTL: 7 * 24 * time.Hour, PhoneRefreshTTL: 10 * 365 * 24 * time.Hour}}}
	near := func(a, b time.Time) bool { d := a.Sub(b); return d < time.Second && d > -time.Second }

	if got := s.signInEnds(context.Background()); !near(got, time.Now().Add(7*24*time.Hour)) {
		t.Fatalf("a new browser sign-in ends %v", got)
	}
	signedIn := time.Now().Add(-6 * 24 * time.Hour).Add(7 * 24 * time.Hour)
	if got := s.signInEnds(keepEnd(context.Background(), signedIn)); !got.Equal(signedIn) {
		t.Fatalf("renewing a browser moved its end to %v", got)
	}
	if got := s.signInEnds(asPhone(keepEnd(context.Background(), signedIn))); !near(got, time.Now().Add(10*365*24*time.Hour)) {
		t.Fatalf("the phone's sign-in ends %v", got)
	}
}

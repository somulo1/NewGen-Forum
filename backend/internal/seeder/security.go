package seeder

import (
	"context"
	"sort"
	"time"

	"hub/internal/app"
	"hub/internal/modules/security"
)

func (s *Seeder) seedSecurity(ctx context.Context) {
	o := s.Owner.P
	// the mock turns rule r7 off; it is on while its history loads so the alert numbers match, then off
	for _, r := range security.Rules {
		s.must(s.Sec.SetRule(ctx, o, r.ID, true, r.Severity), "rule on "+r.ID)
	}

	logs := append([]baseLog(nil), s.data.logs...)
	sort.SliceStable(logs, func(i, j int) bool { return logs[i].Ago > logs[j].Ago }) // oldest first
	// someone has to have signed in from Nairobi before a sign-in from Lagos counts as new
	s.audit(ctx, baseLog{User: "Diana Rotich", Action: "Signed in", Module: "Auth", IP: "41.90.3.10", Loc: "Nairobi, KE", Result: "Success", Ago: 60 * 24 * 7})
	for _, l := range logs {
		s.audit(ctx, l)
	}
	s.must(s.Sec.SetRule(ctx, o, "r7", false, "Low"), "r7 off")

	// the mock rule switches and severities
	var on map[string]bool
	s.st("rulesOn", &on)
	for id, v := range on {
		if id == "r7" {
			continue
		}
		s.must(s.Sec.SetRule(ctx, o, id, v, ruleSeverity(id)), "rule "+id)
	}

	// the alert already being worked in the mock, with its note
	var notes map[string][]struct {
		Status, Comment, Who string
	}
	s.st("alertNotes", &notes)
	alerts, err := s.Sec.Alerts(ctx, o, false)
	s.must(err, "alerts")
	for ref, ns := range notes {
		for _, a := range alerts {
			if a.Ref == ref {
				for _, n := range ns {
					s.must(s.Sec.SetAlertStatus(ctx, o, a.ID, n.Status, n.Comment), "alert note")
				}
			}
		}
	}
}

func ruleSeverity(id string) string {
	for _, r := range security.Rules {
		if r.ID == id {
			return r.Severity
		}
	}
	return "Medium"
}

// audit writes one historic log line through the real audit path (ledger, rules and all), dated as the mock says.
func (s *Seeder) audit(ctx context.Context, l baseLog) {
	at := s.Now.Add(-time.Duration(l.Ago * float64(time.Minute)))
	// two lines are about when something happened, so pin them to the hour the story needs
	switch {
	case l.ID == "g6":
		d := s.Now.In(nairobi).AddDate(0, 0, -1)
		at = time.Date(d.Year(), d.Month(), d.Day(), 23, 40, 0, 0, nairobi).UTC()
	case l.ID == "g4":
		d := s.Now.In(nairobi)
		at = time.Date(d.Year(), d.Month(), d.Day(), 10, 30, 0, 0, nairobi).UTC()
		if at.After(s.Now) {
			at = at.AddDate(0, 0, -1)
		}
	}
	e := &app.Entry{Org: s.OrgID, Name: l.User, Action: l.Action, Module: l.Module, IP: l.IP, Location: l.Loc, Result: l.Result, At: &at}
	if a, ok := s.ByName[l.User]; ok {
		p := a.Person
		e.Actor = &p
	}
	if l.User == "Samwel Okoth Omulo" {
		p := s.Owner.Person
		e.Actor = &p
	}
	s.must(s.A.DB.WithOrg(ctx, s.OrgID, func(tx pgxTx) error { return s.A.AuditEntry(ctx, tx, e) }), "audit "+l.ID)
}

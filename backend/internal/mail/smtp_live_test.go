package mail

import (
	"context"
	"os"
	"testing"
)

// Sends one sample of each kind of email through the configured SMTP server. Skipped unless HUB_SMTP_TEST_TO is set:
//
//	HUB_SMTP_TEST_TO=you@example.com go test ./internal/mail -run Live -v
func TestLiveSMTPSamples(t *testing.T) {
	to := os.Getenv("HUB_SMTP_TEST_TO")
	s := FromEnv()
	if to == "" || s == nil {
		t.Skip("set HUB_SMTP_HOST and HUB_SMTP_TEST_TO to send sample emails")
	}
	samples := []struct {
		subject string
		c       Card
	}{
		{"[Hub sample] You were tagged", Card{Accent: Violet, Org: "Demo Organization", Greeting: "Hello Sam,", Title: "You were tagged", Preheader: "Grace Njeri tagged you",
			Lead: "Grace Njeri tagged you in ticket HELP-214. Open Hub to see what they said.", Rows: [][2]string{{"Tagged by", "Grace Njeri"}, {"Where", "ticket HELP-214"}}, Action: [2]string{"Open in Hub", "http://localhost:3000/tickets"}}},
		{"[Hub sample] HELP-214 was resolved", Card{Accent: Green, Org: "Demo Organization", Greeting: "Hello Sam,", Title: "Your ticket was resolved", Preheader: "Ticket HELP-214 is marked as resolved.",
			Lead: "Good news: the ticket you raised has been resolved. Open it to check the outcome, and tell us if something is still wrong.",
			Rows: [][2]string{{"Ticket", "HELP-214"}, {"Category", "Network & Wi-Fi"}, {"Priority", "High"}, {"Raised by", "Samwel Omulo"}}, Action: [2]string{"Review ticket HELP-214", "http://localhost:3000/tickets"}}},
		{"[Hub sample] HELP-215 was escalated to you", Card{Accent: Red, Org: "Demo Organization", Greeting: "Hello Brian,", Title: "A ticket was escalated to you", Preheader: "Ticket HELP-215 needs a decision.",
			Lead: "This ticket needs someone with more authority to take it forward.", Rows: [][2]string{{"Ticket", "HELP-215"}, {"Subject", "VPN down for Finance"}, {"Category", "Access"}, {"Priority", "High"}, {"Raised by", "Faith Achieng"}, {"Handler", "Brian Kiprop"}, {"Status", "In progress"}, {"Due", "21 Oct 2026, 09:00 UTC"}, {"Attachments", "2 files"}},
			Quote: [2]string{"Why it was escalated", "The VPN has been down for the whole Finance team since 9am."}, Note: "Ticket description: Nobody in Finance can connect since this morning. Error 809 on every laptop.\n\nAttached files are not sent by email. Open the ticket in Hub to see them.", Action: [2]string{"Open ticket HELP-215", "http://localhost:3000/tickets"}}},
		{"[Hub sample] Approval needed", Card{Accent: Amber, Org: "Demo Organization", Greeting: "Hello Brian,", Title: "A request is waiting for your approval", Preheader: "AP-35: Publish the Q4 newsletter",
			Lead: "Grace Njeri sent a request that needs your decision.", Rows: [][2]string{{"Request", "AP-35"}, {"What", "Publish the Q4 newsletter"}, {"Requested by", "Grace Njeri"}, {"Step", "Approval 1 of 2"}},
			Action: [2]string{"Review and decide", "http://localhost:3000/approvals"}, Note: "You can approve or reject it from the Approvals page."}},
		{"[Hub sample] New announcement from Samwel Omulo", Card{Accent: Green, Org: "Demo Organization", Greeting: "Hello Grace,", Title: "New announcement", Preheader: "New announcement from Samwel Omulo",
			Lead: "Samwel Omulo posted an announcement for you.", Rows: [][2]string{{"Type", "Announcement"}, {"From", "Samwel Omulo"}, {"Posted to", "the community wall"}, {"Reference", "C-212"}}, Action: [2]string{"Read it in Hub", "http://localhost:3000/communications"}}},
		{"[Hub sample] Your Hub account was temporarily locked", Card{Greeting: "Hello Sam,", Title: "Your account was temporarily locked", Preheader: "Five wrong passwords were entered",
			Lead: "Someone entered the wrong password five times in a row, so Hub locked your account for a few minutes to keep it safe. It unlocks by itself.",
			Rows: [][2]string{{"Account", "sam@example.com"}, {"Attempts from", "41.90.12.7"}, {"Location", "Nairobi, KE"}, {"Locked until", "12:31 UTC"}, {"Unlocks", "by itself, nothing to do"}},
			Note: "If this was you, wait a few minutes and try again, or reset your password. If it was not you, reset your password now and tell your administrator."}},
		{"[Hub sample] New sign-in to your Hub account", Card{Greeting: "Hello Sam,", Title: "New sign-in to your account", Preheader: "A sign-in from 102.68.4.9",
			Lead: "Your account was just signed in to from an address you have not used before. Signing in here signed you out of your other browsers and devices.",
			Rows: [][2]string{{"When", "20 Oct 2026, 12:20 UTC"}, {"Address", "102.68.4.9"}, {"Location", "Mombasa, KE"}, {"Browser", "Mozilla/5.0 (X11; Linux x86_64) Firefox/131.0"}},
			Note: "If this was you, no action is needed. If it was not you, reset your password now and tell your administrator."}},
		{"[Hub sample] Security alert AL-9: 5 failed sign-ins", Card{Greeting: "Hello Olivia,", Title: "A security alert was raised", Preheader: "AL-9: 5 failed sign-ins for Peter Kamau",
			Lead: "Hub's detection rules flagged something that needs a look.", Rows: [][2]string{{"Alert", "AL-9"}, {"What", "5 failed sign-ins for Peter Kamau"}, {"Severity", "High"}, {"Rule", "r1"}, {"Raised", "20 Oct 2026, 12:21 UTC"}},
			Quote: [2]string{"Details", "Five wrong passwords within 10 minutes from 41.90.12.7. The account was locked for a few minutes."}, Action: [2]string{"Open the alert", "http://localhost:3000/security"}, Note: "Open the alert to investigate it. Every status change needs a comment, except closing a false alarm."}},
		{"[Hub sample] Your verification code", Card{Accent: Blue, Title: "Your verification code", Preheader: "Your code is 482 913", Lead: "Enter this code to continue. If you did not ask for it, you can ignore this message.",
			Rows: [][2]string{{"Your code is", "482 913"}, {"Works for", "10 minutes"}}, Note: "Never share this code with anyone. Hub staff will never ask for it."}},
	}
	for _, m := range samples {
		if err := SendCard(context.Background(), s, to, m.subject, m.c); err != nil {
			t.Fatalf("%s: %v", m.subject, err)
		}
		t.Logf("sent: %s", m.subject)
	}
}

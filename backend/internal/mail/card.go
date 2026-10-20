package mail

import (
	"html"
	"strings"
)

// Card is one system email: a title, a few lines, a table of facts and (optionally) a quoted block and a button. Every message
// Hub sends is built from it, so they all look the same: clean, one column, readable on a phone.
type Card struct {
	Accent    string      // colour of the top bar and the button (default green)
	Preheader string      // the grey preview line shown beside the subject in an inbox
	Org       string      // the organization it concerns, shown small above the title
	Title     string      // what happened, in a few words
	Greeting  string      // "Hello Sam,"
	Lead      string      // one or two sentences
	Rows      [][2]string // facts: label, value
	Quote     [2]string   // optional block: label, text
	Action    [2]string   // optional button: label, URL
	Note      string      // small grey line under the button
}

const (
	Green  = "#2F9E5B"
	Amber  = "#D9822B"
	Red    = "#D64545"
	Blue   = "#2F7DD1"
	Violet = "#7B5CD6"
)

func esc(s string) string { return html.EscapeString(s) }

// HTML renders the card with inline styles (email programs ignore style sheets). It uses the full width of the message: plain
// text on white, with the facts, the quote and the sign-off each in a light grey box.
func (c Card) HTML() string {
	const font = `font-family:Arial,Helvetica,sans-serif;font-size:14px;line-height:1.5;color:#202124`
	box := `margin:0 0 16px;padding:14px 16px;background:#F8F9FA;border:1px solid #E8EAED;border-radius:4px;` + font
	var b strings.Builder
	b.WriteString(`<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><meta name="color-scheme" content="light"><title>` + esc(c.Title) + `</title></head>`)
	b.WriteString(`<body style="margin:0;padding:24px;background:#FFFFFF;` + font + `">`)
	if c.Preheader != "" {
		b.WriteString(`<div style="display:none;max-height:0;overflow:hidden;opacity:0;color:#FFFFFF">` + esc(c.Preheader) + `</div>`)
	}
	if c.Greeting != "" {
		b.WriteString(`<p style="margin:0 0 16px">` + esc(c.Greeting) + `</p>`)
	}
	b.WriteString(`<p style="margin:0 0 16px"><strong>` + esc(c.Title) + `</strong></p>`)
	if c.Lead != "" {
		b.WriteString(`<p style="margin:0 0 16px">` + esc(c.Lead) + `</p>`)
	}
	if len(c.Rows) > 0 {
		b.WriteString(`<div style="` + box + `">`)
		for i, r := range c.Rows {
			gap := "margin:0"
			if i > 0 {
				gap = "margin:6px 0 0"
			}
			b.WriteString(`<div style="` + gap + `"><strong style="color:#5F6368">` + esc(r[0]) + `:</strong> ` + esc(r[1]) + `</div>`)
		}
		b.WriteString(`</div>`)
	}
	if c.Quote[1] != "" {
		b.WriteString(`<div style="` + box + `">`)
		if c.Quote[0] != "" {
			b.WriteString(`<div style="margin:0 0 6px"><strong style="color:#5F6368">` + esc(c.Quote[0]) + `:</strong></div>`)
		}
		b.WriteString(`<div style="white-space:pre-line">` + esc(c.Quote[1]) + `</div></div>`)
	}
	if c.Action[1] != "" {
		b.WriteString(`<p style="margin:0 0 16px"><a href="` + esc(c.Action[1]) + `" style="display:inline-block;background:#1A73E8;color:#FFFFFF;text-decoration:none;font-weight:bold;padding:10px 20px;border-radius:4px">` + esc(c.Action[0]) + `</a></p>`)
	}
	if c.Note != "" {
		b.WriteString(`<p style="margin:0 0 16px;color:#5F6368">` + esc(c.Note) + `</p>`)
	}
	sign := "Hub"
	if c.Org != "" {
		sign = c.Org + " on Hub"
	}
	b.WriteString(`<p style="margin:24px 0 0">Best regards,<br>` + esc(sign) + `</p>`)
	b.WriteString(`<div style="margin:24px -24px -24px;padding:16px 24px;background:#F8F9FA;border-top:1px solid #E8EAED;text-align:center;font-size:12px;color:#80868B">This is an automated notification. Please do not reply to this email.<br>Messages and files in Hub are end-to-end encrypted, so this email never contains them.</div>`)
	b.WriteString(`</body></html>`)
	return b.String()
}

// Text is the plain version, for programs that do not show HTML.
func (c Card) Text() string {
	var b strings.Builder
	b.WriteString(c.Title + "\n\n")
	if c.Greeting != "" {
		b.WriteString(c.Greeting + "\n\n")
	}
	if c.Lead != "" {
		b.WriteString(c.Lead + "\n\n")
	}
	for _, r := range c.Rows {
		b.WriteString(r[0] + ": " + r[1] + "\n")
	}
	if len(c.Rows) > 0 {
		b.WriteString("\n")
	}
	if c.Quote[1] != "" {
		b.WriteString(c.Quote[0] + "\n" + c.Quote[1] + "\n\n")
	}
	if c.Action[1] != "" {
		b.WriteString(c.Action[0] + ": " + c.Action[1] + "\n\n")
	}
	if c.Note != "" {
		b.WriteString(c.Note + "\n\n")
	}
	b.WriteString("Hub\n")
	return b.String()
}

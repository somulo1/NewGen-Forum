package people

import (
	"strings"
	"unicode"
)

// What leaves the API is masked: personal details stay clean in storage and in what people save, and are only partly shown in transit.
const maskDot = "•"

// IsMasked reports a value sent back with mask characters in it. It must never be stored.
func IsMasked(v string) bool { return strings.Contains(v, maskDot) }

// MaskEmail: benta.mbogo@demulla.co.ke -> b•••@d•••.ke
func MaskEmail(v string) string {
	at := strings.LastIndex(v, "@")
	if v == "" {
		return ""
	}
	if at < 1 || at == len(v)-1 {
		return "••••"
	}
	local, dom := []rune(v[:at]), v[at+1:]
	tld := ""
	if dot := strings.LastIndex(dom, "."); dot > 0 {
		tld = dom[dot:]
	}
	return string(local[0]) + "•••@" + string([]rune(dom)[0]) + "•••" + tld
}

// MaskPhone: +254700123456 -> +••• ••• 456
func MaskPhone(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	digits := strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) {
			return r
		}
		return -1
	}, v)
	if len(digits) > 3 {
		digits = digits[len(digits)-3:]
	}
	out := "••• ••• " + digits
	if v[0] == '+' {
		out = "+" + out
	}
	return out
}

// MaskDOB hides the year; the day stays, because birthdays are celebrated in Hub: 1994-03-21 -> ••••-03-21
func MaskDOB(v string) string {
	if len(v) != 10 {
		if v == "" {
			return ""
		}
		return "••••"
	}
	return "••••" + v[4:]
}

// MaskGender hides it entirely.
func MaskGender(v string) string {
	if v == "" {
		return ""
	}
	return "•••"
}

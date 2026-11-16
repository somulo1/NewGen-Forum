package projects

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"

	"hub/internal/httpx"
)

// A task's GitHub branch is stored with the task: made once, from its key and title, when the task is created (or the first time
// an older task is saved), and kept as it is after that, so renaming a task never breaks a branch someone already made. Anyone who
// may edit the task can change it, to match a branch that already exists.

var branchRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._/-]{0,99}$`)

var stopWords = map[string]bool{"a": true, "an": true, "the": true, "and": true, "or": true, "of": true, "to": true, "for": true, "in": true,
	"on": true, "at": true, "with": true, "is": true, "be": true}

// BranchName: "HUB-12" + "Fix the login timeout on mobile!" -> "HUB-12_fix_login_timeout_mobile" (the key, then up to six words of
// the title, lowercase, letters and digits only, joined by underscores, at most 50 characters).
func BranchName(key, title string) string {
	var b strings.Builder
	for _, r := range norm.NFKD.String(strings.ToLower(title)) {
		switch {
		case unicode.Is(unicode.Mn, r):
		case r < 128 && (unicode.IsLetter(r) || unicode.IsDigit(r)):
			b.WriteRune(r)
		default:
			b.WriteByte(' ')
		}
	}
	parts := []string{}
	if k := strings.Trim(regexp.MustCompile(`[^A-Za-z0-9-]+`).ReplaceAllString(key, ""), "-"); k != "" {
		parts = append(parts, k)
	}
	n := 0
	for i, w := range strings.Fields(b.String()) {
		if n == 6 {
			break
		}
		if i > 0 && stopWords[w] {
			continue
		}
		parts = append(parts, w)
		n++
	}
	name := strings.Join(parts, "_")
	for len(name) > 50 {
		i := strings.LastIndex(name[:50], "_")
		if i <= 0 {
			name = name[:50]
			break
		}
		name = name[:i]
	}
	return name
}

// setBranch keeps, checks or makes a task's branch before it is saved.
func setBranch(code string, ref int, data, old map[string]any) error {
	v := strings.TrimSpace(text(data, "branch"))
	switch {
	case v != "":
		if !branchRe.MatchString(v) || strings.Contains(v, "..") || strings.HasSuffix(v, "/") || strings.HasSuffix(v, ".lock") {
			return httpx.Bad("A branch name uses letters, digits and . _ - / (up to 100 characters).")
		}
		data["branch"] = v
	case text(old, "branch") != "":
		data["branch"] = text(old, "branch") // kept as it was made, whatever the title becomes
	default:
		data["branch"] = BranchName(fmt.Sprintf("%s-%d", code, ref), text(data, "title"))
	}
	return nil
}

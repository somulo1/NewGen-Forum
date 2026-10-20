package mail

import (
	"strings"
	"testing"
)

func TestCardEscapesEverythingAndKeepsTheLink(t *testing.T) {
	c := Card{Title: "<b>Hi</b>", Lead: "a & b", Rows: [][2]string{{"x", "<script>"}}, Quote: [2]string{"Why", "line1\nline2"}, Action: [2]string{"Open", "https://hub.example/t?a=1&b=2"}}
	h := c.HTML()
	if strings.Contains(h, "<script>") || strings.Contains(h, "<b>Hi</b>") {
		t.Fatalf("not escaped: %s", h)
	}
	if !strings.Contains(h, `href="https://hub.example/t?a=1&amp;b=2"`) {
		t.Fatalf("link missing: %s", h)
	}
	if !strings.Contains(c.Text(), "https://hub.example/t?a=1&b=2") {
		t.Fatal("plain text has no link")
	}
}

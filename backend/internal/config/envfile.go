package config

import (
	"bufio"
	"log/slog"
	"os"
	"strings"
)

// loadEnvFile reads KEY=VALUE settings from a file when one exists, for platforms where settings are given as a file (Render's
// "Secret Files" appear at /etc/secrets/<name>) and for development (./.env). A setting already in the process environment is
// never overridden. Lines starting with # are comments; values may be wrapped in single or double quotes; "export " is allowed.
func loadEnvFile() {
	paths := []string{os.Getenv("HUB_ENV_FILE"), ".env", "/etc/secrets/.env", "/etc/secrets/env"}
	for _, p := range paths {
		if p == "" {
			continue
		}
		f, err := os.Open(p)
		if err != nil {
			continue
		}
		n := 0
		sc := bufio.NewScanner(f)
		sc.Buffer(make([]byte, 1<<20), 1<<20)
		for sc.Scan() {
			line := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(sc.Text()), "export "))
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			k, v, ok := strings.Cut(line, "=")
			k = strings.TrimSpace(k)
			if !ok || k == "" || strings.ContainsAny(k, " \t") {
				continue
			}
			v = strings.TrimSpace(v)
			if len(v) >= 2 && (v[0] == '"' || v[0] == '\'') {
				if end := strings.IndexByte(v[1:], v[0]); end >= 0 {
					v = v[1 : 1+end]
				}
			} else if i := strings.Index(v, " #"); i >= 0 {
				v = strings.TrimSpace(v[:i])
			}
			if _, set := os.LookupEnv(k); !set {
				_ = os.Setenv(k, v)
				n++
			}
		}
		f.Close()
		slog.Info("settings read from file", "file", p, "settings", n)
	}
}

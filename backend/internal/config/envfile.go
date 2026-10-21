package config

import (
	"bufio"
	"log/slog"
	"os"
	"regexp"
	"strings"
)

// loadEnvFile reads KEY=VALUE settings from a file when one exists, for platforms where settings are given as a file (Render's
// "Secret Files" appear at /etc/secrets/<name>) and for development (./.env). A setting already in the process environment is
// never overridden. Lines starting with # are comments; values may be wrapped in single or double quotes; "export " is allowed.
var secretsDir = "/etc/secrets"

var settingName = regexp.MustCompile(`^HUB_[A-Z0-9_]+$`)

func loadEnvFile() {
	found := false
	named := 0
	paths := []string{os.Getenv("HUB_ENV_FILE"), ".env", "/etc/secrets/.env", "/etc/secrets/env"}
	// Secret Files named like a setting (a file called HUB_ENV holding just `prod`): the file name is the setting and its content is
	// the value. Any other file in the secrets folder whose name mentions "env" (.env.render, prod.env, ...) is read as KEY=VALUE lines.
	if es, err := os.ReadDir(secretsDir); err == nil {
		for _, e := range es {
			if e.IsDir() {
				continue
			}
			if settingName.MatchString(e.Name()) {
				if b, err := os.ReadFile(secretsDir + "/" + e.Name()); err == nil {
					if _, set := os.LookupEnv(e.Name()); !set {
						_ = os.Setenv(e.Name(), strings.TrimSpace(string(b)))
						found = true
						named++
					}
				}
				continue
			}
			if strings.Contains(strings.ToLower(e.Name()), "env") {
				paths = append(paths, secretsDir+"/"+e.Name())
			}
		}
	}
	defer func() {
		if named > 0 {
			slog.Info("settings read from secret files named after them", "settings", named)
		}
		if !found && os.Getenv("RENDER") != "" {
			var names []string
			if es, err := os.ReadDir(secretsDir); err == nil {
				for _, e := range es {
					names = append(names, e.Name())
				}
			}
			slog.Warn("no settings file found", "files in /etc/secrets", names, "looked for", []string{".env", "/etc/secrets/.env"})
		}
	}()
	for _, p := range paths {
		if p == "" {
			continue
		}
		f, err := os.Open(p)
		if err != nil {
			continue
		}
		found = true
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

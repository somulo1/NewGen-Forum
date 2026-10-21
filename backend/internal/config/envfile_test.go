package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnvFileFillsOnlyMissingSettings(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "settings")
	_ = os.WriteFile(f, []byte("# comment\nTEST_A=one\nexport TEST_B='two words'\nTEST_C=\"3\"  \nTEST_D=four # note\nTEST_KEEP=file\n\nbad line\n"), 0o600)
	t.Setenv("HUB_ENV_FILE", f)
	t.Setenv("TEST_KEEP", "process")
	for _, k := range []string{"TEST_A", "TEST_B", "TEST_C", "TEST_D"} {
		os.Unsetenv(k)
	}
	loadEnvFile()
	got := map[string]string{"TEST_A": "one", "TEST_B": "two words", "TEST_C": "3", "TEST_D": "four", "TEST_KEEP": "process"}
	for k, want := range got {
		if os.Getenv(k) != want {
			t.Errorf("%s = %q, want %q", k, os.Getenv(k), want)
		}
	}
}

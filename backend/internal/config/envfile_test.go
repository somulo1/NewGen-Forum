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

// Render's Secret Files can be made one per setting: the file is named HUB_ENV and holds just the value.
func TestSecretFilesNamedAfterSettings(t *testing.T) {
	dir := t.TempDir()
	old := secretsDir
	secretsDir = dir
	defer func() { secretsDir = old }()
	_ = os.WriteFile(filepath.Join(dir, "HUB_TEST_ONE"), []byte("prod\n"), 0o600)
	_ = os.WriteFile(filepath.Join(dir, "HUB_TEST_TWO"), []byte(" postgres://u:p@h/db?sslmode=require \n"), 0o600)
	_ = os.WriteFile(filepath.Join(dir, "HUB_TEST_SET"), []byte("file"), 0o600)
	os.Unsetenv("HUB_TEST_ONE")
	os.Unsetenv("HUB_TEST_TWO")
	t.Setenv("HUB_TEST_SET", "process")
	t.Setenv("HUB_ENV_FILE", "")
	loadEnvFile()
	if os.Getenv("HUB_TEST_ONE") != "prod" || os.Getenv("HUB_TEST_TWO") != "postgres://u:p@h/db?sslmode=require" || os.Getenv("HUB_TEST_SET") != "process" {
		t.Fatalf("one=%q two=%q set=%q", os.Getenv("HUB_TEST_ONE"), os.Getenv("HUB_TEST_TWO"), os.Getenv("HUB_TEST_SET"))
	}
}

func TestRedisAddressAcceptsProviderURLs(t *testing.T) {
	for in, want := range map[string][3]string{
		"127.0.0.1:6390":                      {"127.0.0.1:6390", "", "false"},
		"redis://red-abc:6379":                {"red-abc:6379", "", "false"},
		"redis://default:s3cret@red-abc:6379": {"red-abc:6379", "s3cret", "false"},
		"rediss://:pw@host.example.com:6380":  {"host.example.com:6380", "pw", "true"},
		"redis://red-abc":                     {"red-abc:6379", "", "false"},
	} {
		t.Setenv("HUB_REDIS_ADDR", in)
		t.Setenv("HUB_REDIS_PASS", "")
		c, err := Load()
		if err != nil {
			t.Fatal(err)
		}
		got := [3]string{c.RedisAddr, c.RedisPass, map[bool]string{true: "true", false: "false"}[c.RedisTLS]}
		if got != want {
			t.Errorf("%q -> %v, want %v", in, got, want)
		}
	}
}

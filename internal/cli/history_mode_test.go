package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// TestHistoryModeHonoursConfigPath pins the round-053 (ADR 0022) fix: an offline
// session command's mode comes from the `-c` configuration's MODE when the env
// override is unset; the env still wins; an explicit unreadable `-c` is an error;
// an absent default degrades to "butler".
func TestHistoryModeHonoursConfigPath(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "configs"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := filepath.Join(dir, "configs", "a.yaml")
	if err := os.WriteFile(cfg, []byte("MODE: alpha\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("TELL_ME_MODE", "")

	got, err := historyMode(dir, cfg)
	if err != nil || got != "alpha" {
		t.Fatalf("historyMode(-c a.yaml) = %q, %v; want alpha (the config's MODE)", got, err)
	}

	// The env override still wins over -c.
	t.Setenv("TELL_ME_MODE", "gamma")
	got, err = historyMode(dir, cfg)
	if err != nil || got != "gamma" {
		t.Fatalf("historyMode with TELL_ME_MODE set = %q, %v; want gamma (env wins)", got, err)
	}
	t.Setenv("TELL_ME_MODE", "")

	// An explicit `-c` that cannot be read is a hard error (Q2 → (A)).
	if _, err := historyMode(dir, filepath.Join(dir, "configs", "missing.yaml")); err == nil {
		t.Fatal("an explicit missing -c must be an error, not a silent fallback")
	}

	// No -c and no default config degrades to butler (round-007 tolerance).
	got, err = historyMode(dir, "")
	if err != nil || got != "butler" {
		t.Fatalf("historyMode(no -c, no default) = %q, %v; want butler, nil", got, err)
	}
}

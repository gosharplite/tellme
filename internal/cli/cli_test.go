package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// T003 materialised as a real test — the CLI package's one pure helper (default
// path discovery) is covered here (research Decision 1 & 3).
func TestDefaultConfigPath(t *testing.T) {
	tests := []struct {
		name  string
		mode  string
		unset bool
		want  string
	}{
		{name: "mode taken from TELL_ME_MODE", mode: "coder", want: filepath.Join("/home/x", "configs", "coder.yaml")},
		{name: "defaults to butler when unset", unset: true, want: filepath.Join("/home/x", "configs", "butler.yaml")},
		{name: "defaults to butler when blank", mode: "", want: filepath.Join("/home/x", "configs", "butler.yaml")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TELL_ME_MODE", "placeholder")
			if tt.unset {
				if err := os.Unsetenv("TELL_ME_MODE"); err != nil {
					t.Fatalf("unset TELL_ME_MODE: %v", err)
				}
			} else {
				t.Setenv("TELL_ME_MODE", tt.mode)
			}
			if got := defaultConfigPath("/home/x"); got != tt.want {
				t.Fatalf("defaultConfigPath = %q, want %q", got, tt.want)
			}
		})
	}
}

// Review finding #1/#2/#3 regression tests for resolve().

func writeConfigFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write config %q: %v", path, err)
	}
}

// TestResolveRejectsEmptyAfterExpansion pins review finding #2: a mandatory field
// whose placeholder resolves to empty must fail on the RESOLVED state
// (expand-then-validate), not slip past the non-empty invariant.
func TestResolveRejectsEmptyAfterExpansion(t *testing.T) {
	// Neutralise ambient TELL_ME_* overrides so the file's provider is selected.
	t.Setenv("TELL_ME_SELECTED_PROVIDER", "")
	t.Setenv("TELL_ME_MODE", "butler")
	// The referenced variable is set-but-empty: expansion substitutes the empty default.
	t.Setenv("TELLME_TEST_EMPTY_URL", "")

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	writeConfigFile(t, cfgPath, "MODE: butler\n"+
		"SELECTED_PROVIDER: prov\n"+
		"PROVIDERS:\n"+
		"  prov:\n"+
		"    TYPE: openai\n"+
		"    MODEL: gpt-5.5\n"+
		"    URL: ${TELLME_TEST_EMPTY_URL:-}\n")

	_, rerr := resolve(dir, cfgPath)
	if rerr == nil {
		t.Fatal("resolve() = nil error, want provider-invalid (URL empty after expansion)")
	}
	if rerr.Reason != reasonProviderInvalid {
		t.Fatalf("resolve() reason = %q, want %q", rerr.Reason, reasonProviderInvalid)
	}
}

// TestResolveCarriesExpandedProvider pins review finding #3: resolve() carries
// the resolved (expanded) selected provider so Slice 004 need not re-load the
// configuration.
func TestResolveCarriesExpandedProvider(t *testing.T) {
	t.Setenv("TELL_ME_SELECTED_PROVIDER", "")
	t.Setenv("TELL_ME_MODE", "butler")
	t.Setenv("TELLME_TEST_KEY", "secret-xyz")

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	writeConfigFile(t, cfgPath, "MODE: butler\n"+
		"SELECTED_PROVIDER: prov\n"+
		"PROVIDERS:\n"+
		"  prov:\n"+
		"    TYPE: openai\n"+
		"    MODEL: gpt-5.5\n"+
		"    URL: https://example.test/v1\n"+
		"    API_KEY: ${TELLME_TEST_KEY}\n")

	res, rerr := resolve(dir, cfgPath)
	if rerr != nil {
		t.Fatalf("resolve() unexpected error: %v", rerr)
	}
	if res.Selected != "prov" {
		t.Errorf("Selected = %q, want prov", res.Selected)
	}
	if res.Provider.Model != "gpt-5.5" {
		t.Errorf("carried Model = %q, want gpt-5.5", res.Provider.Model)
	}
	if res.Provider.APIKey != "secret-xyz" {
		t.Errorf("carried APIKey = %q, want secret-xyz (expansion not applied to carried provider)", res.Provider.APIKey)
	}
}

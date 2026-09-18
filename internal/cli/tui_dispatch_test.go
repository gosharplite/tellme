package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestTUIDispatchEngagesWhenEnabledOnTerminal (round-015 T030; round-016 T024):
// with -i and a terminal stdin, run() delegates to the tuiPromptRunner seam — the
// flag/TTY gating is unit-testable without a terminal event loop (PR #38 review
// directive ④). Round 016 dropped the obsolete store param (architect D3).
func TestTUIDispatchEngagesWhenEnabledOnTerminal(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, "configs"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := "MODE: butler\nPERSON: p\nSELECTED_PROVIDER: m\nPROVIDERS:\n  m:\n    TYPE: deepseek\n    MODEL: x\n    URL: https://example.invalid\n"
	if err := os.WriteFile(filepath.Join(home, "configs", "butler.yaml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TELL_ME_HOME", home)

	called := false
	opts := Options{Deps: defaultTestDeps(), Prompter: fakePrompter{called: &called}}

	var out, errOut strings.Builder
	e := runtimeEnv{
		stdin:    strings.NewReader(""),
		stdout:   &out,
		stderr:   &errOut,
		isTTY:    func(any) bool { return true },
		renderer: &stubRenderer{},
	}
	_ = run([]string{"-i"}, "dev", opts, e)
	if !called {
		t.Fatalf("the interactive prompt did not engage on a terminal with -i; stderr=%q", errOut.String())
	}
}

// TestTUIDispatchFallsBackOnNonTerminal (round-015 T030): with -i but a
// NON-terminal stdin, the seam is NOT invoked (the piped/plain path applies).
func TestTUIDispatchFallsBackOnNonTerminal(t *testing.T) {
	// Hermeticity (round-015 PR #38 review): a non-terminal `-i` run reads the
	// piped prompt and routes to the turn path. Without a cleared environment the
	// ambient TELL_ME_HOME could resolve a real provider and dial the network.
	// Neutralize the ambient overrides and force the home unset so the turn stops
	// at home-unset (no network).
	clearAmbientOverrides(t)
	t.Setenv("TELL_ME_HOME", "")

	called := false
	opts := Options{Deps: defaultTestDeps(), Prompter: fakePrompter{called: &called}}

	var out, errOut strings.Builder
	e := runtimeEnv{
		stdin:    strings.NewReader("piped prompt"),
		stdout:   &out,
		stderr:   &errOut,
		isTTY:    func(any) bool { return false },
		renderer: &stubRenderer{},
	}
	_ = run([]string{"-i"}, "dev", opts, e)
	if called {
		t.Fatalf("the interactive prompt engaged on a non-terminal stdin; stderr=%q", errOut.String())
	}
}

// TestTUIDispatchFailsLoudlyWithoutPrompter (round 048 / ADR 0017): with -i on a
// terminal but no Prompter injected, the run fails loudly with the environment
// class (never a silent no-op) — the composition root always injects the port,
// so this is a defensive path.
func TestTUIDispatchFailsLoudlyWithoutPrompter(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, "configs"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := "MODE: butler\nPERSON: p\nSELECTED_PROVIDER: m\nPROVIDERS:\n  m:\n    TYPE: deepseek\n    MODEL: x\n    URL: https://example.invalid\n"
	if err := os.WriteFile(filepath.Join(home, "configs", "butler.yaml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TELL_ME_HOME", home)

	opts := Options{Deps: defaultTestDeps()} // Prompter nil

	var out, errOut strings.Builder
	e := runtimeEnv{
		stdin:    strings.NewReader(""),
		stdout:   &out,
		stderr:   &errOut,
		isTTY:    func(any) bool { return true },
		renderer: &stubRenderer{},
	}
	if code := run([]string{"-i"}, "dev", opts, e); code != EnvironmentError {
		t.Fatalf("code = %d, want EnvironmentError; stderr=%q", code, errOut.String())
	}
}

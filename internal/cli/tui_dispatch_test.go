package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gosharplite/tellme/internal/domain/history"
)

// TestTUIDispatchEngagesWhenEnabledOnTerminal (round-015 T030): with -i and a
// terminal stdin, run() delegates to the tuiPromptRunner seam — the flag/TTY
// gating is unit-testable without a terminal event loop (PR #38 review
// directive ④).
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

	orig := newTUIPromptRunner
	defer func() { newTUIPromptRunner = orig }()
	called := false
	newTUIPromptRunner = func(_ context.Context, _ resolution, _ history.Store, _ runtimeEnv) (string, bool, error) {
		called = true
		return "", false, nil
	}

	var out, errOut strings.Builder
	e := runtimeEnv{
		stdin:    strings.NewReader(""),
		stdout:   &out,
		stderr:   &errOut,
		isTTY:    func(any) bool { return true },
		renderer: &stubRenderer{},
	}
	_ = run([]string{"-i"}, "dev", e)
	if !called {
		t.Fatalf("the interactive prompt did not engage on a terminal with -i; stderr=%q", errOut.String())
	}
}

// TestTUIDispatchFallsBackOnNonTerminal (round-015 T030): with -i but a
// NON-terminal stdin, the seam is NOT invoked (the piped/plain path applies).
func TestTUIDispatchFallsBackOnNonTerminal(t *testing.T) {
	orig := newTUIPromptRunner
	defer func() { newTUIPromptRunner = orig }()
	called := false
	newTUIPromptRunner = func(_ context.Context, _ resolution, _ history.Store, _ runtimeEnv) (string, bool, error) {
		called = true
		return "", false, nil
	}

	var out, errOut strings.Builder
	e := runtimeEnv{
		stdin:    strings.NewReader("piped prompt"),
		stdout:   &out,
		stderr:   &errOut,
		isTTY:    func(any) bool { return false },
		renderer: &stubRenderer{},
	}
	_ = run([]string{"-i"}, "dev", e)
	if called {
		t.Fatalf("the interactive prompt engaged on a non-terminal stdin; stderr=%q", errOut.String())
	}
}

package cli

import (
	"bytes"
	"io"
	"path/filepath"
	"testing"

	"github.com/gosharplite/tellme/internal/domain/history"
	"github.com/gosharplite/tellme/internal/domain/render"
)

// Round 073 (ADR 0045) — the `-l` listing's CLI wiring.

// TestRenderHistoryListProjectsAndGates pins that renderHistoryList projects the
// last N messages as role-tagged listing messages, passes the stdout-gated
// colour, the raw flag, and the resolved width to the port.
func TestRenderHistoryListProjectsAndGates(t *testing.T) {
	t.Setenv("TELL_ME_MODE", "butler")
	t.Setenv("TELL_ME_HOME", t.TempDir())
	home := t.TempDir()

	entries := []history.Entry{
		{Prompt: "q1", Answer: "a1"},
		{Prompt: "q2", Answer: "a2"},
	}
	store := func(string) history.Store { return &fakeStore{entries: entries} }
	var got *fakeListing
	var out, errOut bytes.Buffer
	env := runtimeEnv{
		stdout:    &out,
		stderr:    &errOut,
		stdoutTTY: func(any) bool { return true },
	}
	mk := func() render.Listing { got = &fakeListing{}; return got }

	if code := renderHistoryList(home, 3, "", false, env, store, mk); code != Success {
		t.Fatalf("code = %d, want Success", code)
	}
	if len(got.got) != 3 {
		t.Fatalf("projected %d messages, want 3 (last 3)", len(got.got))
	}
	// The last 3 of [op q1, model a1, op q2, model a2] = [model a1, op q2, model a2].
	wantRoles := []render.ListingRole{render.ListingModel, render.ListingOperator, render.ListingModel}
	for i, want := range wantRoles {
		if got.got[i].Role != want {
			t.Errorf("message[%d].Role = %v, want %v", i, got.got[i].Role, want)
		}
	}
	if !got.spec.Colour {
		t.Error("spec.Colour = false, want true (the stdout probe reports a terminal)")
	}
	if got.spec.Raw {
		t.Error("spec.Raw = true, want false")
	}
	if got.spec.Warn != io.Writer(&errOut) {
		t.Error("spec.Warn must be the diagnostic stream")
	}
}

// TestRenderHistoryListRawAndNoTerminal pins that -r is threaded and that a
// non-terminal stdout leaves the accent off.
func TestRenderHistoryListRawAndNoTerminal(t *testing.T) {
	home := t.TempDir()
	t.Setenv("TELL_ME_MODE", "butler")
	store := func(string) history.Store { return &fakeStore{entries: []history.Entry{{Prompt: "q", Answer: "a"}}} }
	var got *fakeListing
	var out, errOut bytes.Buffer
	env := runtimeEnv{stdout: &out, stderr: &errOut} // no stdout probe ⇒ not a terminal
	mk := func() render.Listing { got = &fakeListing{}; return got }
	if code := renderHistoryList(home, 2, "", true, env, store, mk); code != Success {
		t.Fatalf("code = %d, want Success", code)
	}
	if got.spec.Colour {
		t.Error("spec.Colour = true, want false (no stdout terminal probe)")
	}
	if !got.spec.Raw {
		t.Error("spec.Raw = false, want true")
	}
}

// TestRenderHistoryListWidthIsBestEffort pins research D7: an unreadable config
// degrades to width 0 (the renderer default) instead of failing the listing,
// while a readable WRAP_WIDTH is honoured.
func TestRenderHistoryListWidthIsBestEffort(t *testing.T) {
	home := t.TempDir()
	t.Setenv("TELL_ME_MODE", "butler")
	store := func(string) history.Store { return &fakeStore{} }
	var got *fakeListing
	var out, errOut bytes.Buffer
	env := runtimeEnv{stdout: &out, stderr: &errOut}
	mk := func() render.Listing { got = &fakeListing{}; return got }

	// Unreadable explicit config ⇒ width 0, still Success.
	if code := renderHistoryList(home, 1, filepath.Join(home, "missing.yaml"), false, env, store, mk); code != Success {
		t.Fatalf("code = %d, want Success on an unreadable config", code)
	}
	if got.spec.Width != 0 {
		t.Errorf("spec.Width = %d, want 0 (best-effort degrade)", got.spec.Width)
	}

	// A readable config with WRAP_WIDTH ⇒ honoured.
	cfg := filepath.Join(home, "cfg.yaml")
	writeConfigFile(t, cfg, "MODE: butler\nWRAP_WIDTH: 64\nSELECTED_PROVIDER: p\nPROVIDERS:\n  p:\n    TYPE: openai\n    MODEL: m\n    URL: https://example.test/v1\n")
	t.Setenv("TELL_ME_WRAP_WIDTH", "")
	if code := renderHistoryList(home, 1, cfg, false, env, store, mk); code != Success {
		t.Fatalf("code = %d, want Success", code)
	}
	if got.spec.Width != 64 {
		t.Errorf("spec.Width = %d, want 64 (the resolved WRAP_WIDTH)", got.spec.Width)
	}
}

// TestStdoutTerminalSeam pins the round-073 stdout probe + its hermetic seam.
func TestStdoutTerminalSeam(t *testing.T) {
	t.Setenv("TELL_ME_FORCE_STDOUT_TTY", "1")
	if probe := stdoutTerminalDetector(); !probe(nil) {
		t.Error("TELL_ME_FORCE_STDOUT_TTY=1 must force the stdout probe to report a terminal")
	}
	t.Setenv("TELL_ME_FORCE_STDOUT_TTY", "")
	if env := (runtimeEnv{stdout: io.Discard}); env.stdoutIsTerminal() {
		t.Error("a nil stdout probe must report not-a-terminal (the safe default)")
	}
	if env := (runtimeEnv{stdout: io.Discard, stdoutTTY: func(any) bool { return true }}); !env.stdoutIsTerminal() {
		t.Error("an injected stdout probe must be consulted")
	}
}

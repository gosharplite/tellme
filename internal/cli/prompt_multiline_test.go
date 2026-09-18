package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Round-012 interactive multi-line reader unit tests. The pty-less E2E harness
// cannot drive a terminal branch, so the positive read + the empty/cancel
// contract are asserted here through the injected terminal seam (research
// Decision 6).

func TestReadInteractivePrompt_PrintsHintAndReadsToEOF(t *testing.T) {
	var stderr bytes.Buffer
	got, ok := readInteractivePrompt(context.Background(),
		strings.NewReader("Summarise this:\nthe launch code is ORANGE\n"), &stderr)
	if !ok {
		t.Fatalf("readInteractivePrompt ok = false, want true")
	}
	if got != "Summarise this:\nthe launch code is ORANGE" {
		t.Errorf("captured = %q, want the trimmed multi-line prompt", got)
	}
	if !strings.Contains(stderr.String(), "Reading multi-line input") {
		t.Errorf("stderr = %q, want the multi-line hint", stderr.String())
	}
}

func TestReadInteractivePrompt_EmptySendsNothing(t *testing.T) {
	var stderr bytes.Buffer
	got, ok := readInteractivePrompt(context.Background(), strings.NewReader(""), &stderr)
	if !ok || got != "" {
		t.Errorf("empty read = (%q,%v), want (\"\",true)", got, ok)
	}
}

func TestReadInteractivePrompt_BoundsInput(t *testing.T) {
	huge := strings.Repeat("a", maxStdinBytes+4096)
	got, ok := readInteractivePrompt(context.Background(), strings.NewReader(huge), &bytes.Buffer{})
	if !ok {
		t.Fatalf("ok = false, want true")
	}
	if len(got) > maxStdinBytes {
		t.Errorf("captured %d bytes, want <= %d", len(got), maxStdinBytes)
	}
}

func TestReadInteractivePrompt_CancelReturnsNotOK(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancelled before the read
	got, ok := readInteractivePrompt(ctx, blockingReader{}, &bytes.Buffer{})
	if ok || got != "" {
		t.Errorf("cancelled read = (%q,%v), want (\"\",false)", got, ok)
	}
}

// blockingReader never returns; a cancelled context must short-circuit the read.
type blockingReader struct{}

func (blockingReader) Read([]byte) (int, error) { select {} }

// TestRun_InteractiveEmptyExitsSuccessfully pins the dispatch: a bare invocation
// on a terminal with an empty submission sends no request and exits successfully
// (no configuration is required).
func TestRun_InteractiveEmptyExitsSuccessfully(t *testing.T) {
	var out, errOut bytes.Buffer
	env := runtimeEnv{stdin: strings.NewReader(""), stdout: &out, stderr: &errOut,
		isTTY: func(any) bool { return true }, renderer: &stubRenderer{}}
	if code := run(nil, "dev", testOptions(), env); code != Success {
		t.Fatalf("run(...) = %d, want Success", code)
	}
	if !strings.Contains(errOut.String(), "Reading multi-line input") {
		t.Errorf("stderr = %q, want the multi-line hint", errOut.String())
	}
	if out.Len() != 0 {
		t.Errorf("stdout = %q, want empty", out.String())
	}
}

// TestRun_InteractivePromptRoutesToTurn pins that a non-empty interactive read is
// routed to the reasoning-turn path (not boot / success). With no runtime home
// the turn's resolve fails with the environment class phrase.
func TestRun_InteractivePromptRoutesToTurn(t *testing.T) {
	t.Setenv("TELL_ME_HOME", "")
	var out, errOut bytes.Buffer
	env := runtimeEnv{stdin: strings.NewReader("hello\n"), stdout: &out, stderr: &errOut,
		isTTY: func(any) bool { return true }, renderer: &stubRenderer{}}
	if code := run(nil, "dev", testOptions(), env); code != EnvironmentError {
		t.Fatalf("run(...) = %d, want EnvironmentError (routed to the turn path)", code)
	}
}

// TestDefaultIsTerminalRejectsNullDevice pins the round-012 review BLOCKER B1
// fix: a character device that is NOT a terminal (the null device) must not be
// reported as a terminal. A bare os.ModeCharDevice test returned true here and
// silently masked a missing-configuration failure as exit 0.
func TestDefaultIsTerminalRejectsNullDevice(t *testing.T) {
	f, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatalf("open %s: %v", os.DevNull, err)
	}
	defer func() { _ = f.Close() }()
	if defaultIsTerminal(f) {
		t.Errorf("defaultIsTerminal(%s) = true, want false (a char device is not a terminal)", os.DevNull)
	}
}

// TestDefaultIsTerminalRejectsPipe pins that a pipe is not a terminal (the
// round-005 piped-input path depends on the probe being false here).
func TestDefaultIsTerminalRejectsPipe(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	defer func() { _ = r.Close() }()
	defer func() { _ = w.Close() }()
	if defaultIsTerminal(r) {
		t.Error("defaultIsTerminal(pipe) = true, want false")
	}
}

// TestMultiLineHintMatchesDSLLiteral guards the single-sourced hint against
// drift from the DSL truth (round-012 review TD1): the exported literal must
// equal the exact string pinned in chat/dsl.md. If either side changes alone,
// this fails.
func TestMultiLineHintMatchesDSLLiteral(t *testing.T) {
	const dslLiteral = "[Reading multi-line input. Press Ctrl+C to cancel, or Ctrl+D to send]"
	if MultiLineHint != dslLiteral {
		t.Errorf("MultiLineHint = %q, want the DSL literal %q", MultiLineHint, dslLiteral)
	}
}

// clearAmbientOverrides neutralizes the ambient TELL_ME_*/MAX_* overrides the
// project shell exports (e.g. TELL_ME_MODE), so a unit test resolves a
// deterministic default mode. It mirrors the E2E harness's beforeScenario
// hermeticity (round-009); without it, a test that asserts a workspace path
// fails wherever TELL_ME_MODE selects a different mode.
func clearAmbientOverrides(t *testing.T) {
	t.Helper()
	for _, name := range []string{
		"TELL_ME_MODE",
		"TELL_ME_SELECTED_PROVIDER",
		"TELL_ME_WRAP_WIDTH",
		"MAX_TOOL_LOOP",
		"MAX_HISTORY_TOKENS",
		"TELL_ME_FORCE_STDIN_TTY",
	} {
		t.Setenv(name, "")
	}
}

// TestRun_NewInteractivePrintsHintThenRoutesToTurn pins amendment A8: a
// prompt-less `--new` on a terminal archives the session, then engages the
// reader; a non-empty read is routed to the turn path.
func TestRun_NewInteractivePrintsHintThenRoutesToTurn(t *testing.T) {
	clearAmbientOverrides(t)
	t.Setenv("TELL_ME_HOME", t.TempDir()) // home present so the archive step succeeds
	var out, errOut bytes.Buffer
	env := runtimeEnv{stdin: strings.NewReader("hello\n"), stdout: &out, stderr: &errOut,
		isTTY: func(any) bool { return true }, renderer: &stubRenderer{}}
	code := run([]string{"--new"}, "dev", testOptions(), env)
	if code != ConfigError {
		t.Fatalf("run(--new, tty) = %d, want ConfigError (routed to the turn path; no config)", code)
	}
	if !strings.Contains(errOut.String(), "Reading multi-line input") {
		t.Errorf("stderr = %q, want the multi-line hint", errOut.String())
	}
}

// TestRun_NewInteractiveEmptyArchivesAndSucceeds pins the empty/cancel contract
// for a prompt-less `--new` on a terminal: the session is archived, no request is
// made, and the run exits success.
func TestRun_NewInteractiveEmptyArchivesAndSucceeds(t *testing.T) {
	clearAmbientOverrides(t)
	home := t.TempDir()
	t.Setenv("TELL_ME_HOME", home)
	var out, errOut bytes.Buffer
	env := runtimeEnv{stdin: strings.NewReader(""), stdout: &out, stderr: &errOut,
		isTTY: func(any) bool { return true }, renderer: &stubRenderer{}}
	if code := run([]string{"--new"}, "dev", testOptions(), env); code != Success {
		t.Fatalf("run(--new, tty, empty) = %d, want Success", code)
	}
	if !strings.Contains(errOut.String(), "Reading multi-line input") {
		t.Errorf("stderr = %q, want the multi-line hint", errOut.String())
	}
	if out.Len() != 0 {
		t.Errorf("stdout = %q, want empty", out.String())
	}
	// The archive step resolved/prepared the workspace (default mode "butler").
	if _, err := os.Stat(filepath.Join(home, "output", "butler")); err != nil {
		t.Errorf("workspace not prepared for the fresh session: %v", err)
	}
}

// TestRun_NewNonTTYDoesNotRead pins that a prompt-less `--new` on a NON-terminal
// (empty pipe) keeps its round-007 behaviour: archive and exit, with no
// reader/hint.
func TestRun_NewNonTTYDoesNotRead(t *testing.T) {
	clearAmbientOverrides(t)
	t.Setenv("TELL_ME_HOME", t.TempDir())
	var out, errOut bytes.Buffer
	env := runtimeEnv{stdin: strings.NewReader(""), stdout: &out, stderr: &errOut,
		isTTY: func(any) bool { return false }, renderer: &stubRenderer{}}
	if code := run([]string{"--new"}, "dev", testOptions(), env); code != Success {
		t.Fatalf("run(--new, non-tty) = %d, want Success", code)
	}
	if strings.Contains(errOut.String(), "Reading multi-line input") {
		t.Errorf("stderr = %q, want no reading hint on a non-terminal", errOut.String())
	}
}

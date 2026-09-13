package cli

import (
	"bytes"
	"context"
	"os"
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
	if code := run(nil, "dev", env); code != Success {
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
	if code := run(nil, "dev", env); code != EnvironmentError {
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

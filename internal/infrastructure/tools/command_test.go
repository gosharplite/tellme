package tools

import (
	"context"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// T034/T035 — the command tool, its timeout coverage, and the reader FR-018 path.

// runWithWait executes a command tool call with a hard test-side deadline so a
// regression that wedges fails loudly instead of hanging the suite.
func runWithWait(t *testing.T, ctx context.Context, args string, wait time.Duration) string {
	t.Helper()
	type res struct {
		out string
		err error
	}
	ch := make(chan res, 1)
	go func() {
		o, e := executeCommand{}.Execute(ctx, args, testBudget)
		ch <- res{o, e}
	}()
	select {
	case r := <-ch:
		if r.err != nil {
			t.Fatalf("execute_command: %v", r.err)
		}
		return r.out
	case <-time.After(wait):
		t.Fatalf("execute_command wedged (no result within %v)", wait)
		return ""
	}
}

func TestExecuteCommandNonZeroIsSuccessResult(t *testing.T) {
	got, err := executeCommand{}.Execute(context.Background(), `{"command":"exit 3","reason":"r"}`, testBudget)
	if err != nil {
		t.Fatalf("a non-zero exit must be a successful result, got error %v", err)
	}
	if !strings.Contains(got, "Exit Code: 3") {
		t.Fatalf("result %q does not carry the exit status", got)
	}
}

func TestExecuteCommandSuccessResult(t *testing.T) {
	got, err := executeCommand{}.Execute(context.Background(), `{"command":"echo hi","reason":"r"}`, testBudget)
	if err != nil {
		t.Fatalf("execute_command: %v", err)
	}
	if !strings.Contains(got, "hi") || !strings.Contains(got, "Exit Code: 0") {
		t.Fatalf("result %q; want the output and Exit Code: 0", got)
	}
}

func TestExecuteCommandMissingCommandErrors(t *testing.T) {
	if _, err := (executeCommand{}).Execute(context.Background(), `{"reason":"r"}`, testBudget); err == nil {
		t.Error("a missing command must error")
	}
}

func TestExecuteCommandContractDefault(t *testing.T) {
	if got := (executeCommand{}).Contract().DefaultTimeout; got != commandDefaultTimeout {
		t.Errorf("command default timeout = %v; want %v", got, commandDefaultTimeout)
	}
}

func TestExecuteCommandObservedDeadlineIsNilError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	got := runWithWait(t, ctx, `{"command":"sleep 60","reason":"r"}`, 5*time.Second)
	if !strings.Contains(got, "stopped at the time limit") {
		t.Fatalf("result %q does not record a stop", got)
	}
}

// TestExecuteCommandOutputFileRespectsTimeout witnesses review B1: a hanging
// command with output_file set still returns the stop marker within the deadline.
func TestExecuteCommandOutputFileRespectsTimeout(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out.txt")
	args := `{"command":"sleep 60","output_file":"` + out + `","reason":"r"}`
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	got := runWithWait(t, ctx, args, 5*time.Second)
	if !strings.Contains(got, "stopped at the time limit") {
		t.Fatalf("result %q does not record a stop", got)
	}
}

// TestExecuteCommandGroupEscapeDoesNotWedge witnesses review B2: a descendant
// that escapes the process group (setsid) cannot wedge the drain, because the
// capture path closes the pipe read ends after the kill.
func TestExecuteCommandGroupEscapeDoesNotWedge(t *testing.T) {
	if _, err := exec.LookPath("setsid"); err != nil {
		t.Skip("setsid is not available on this host")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	// setsid puts the sleep in a NEW session/group, so kill(-pgid) misses it; it
	// still holds the inherited stdout pipe. Without closing the read ends the
	// drain would block forever.
	got := runWithWait(t, ctx, `{"command":"setsid sleep 60 & wait","reason":"r"}`, 8*time.Second)
	if !strings.Contains(got, "stopped at the time limit") {
		t.Fatalf("result %q does not record a stop", got)
	}
}

func TestReaderObservedDeadlineIsNilErrorTimeoutResult(t *testing.T) {
	// A deadline already in the past (deterministic — no sleep).
	ctx, cancel := context.WithDeadline(context.Background(), time.Unix(0, 0))
	defer cancel()
	got, err := listFiles{}.Execute(ctx, `{"reason":"r"}`, testBudget)
	if err != nil {
		t.Fatalf("a reader deadline must be a nil-error result, got %v", err)
	}
	if !strings.Contains(got, "stopped at the time limit") {
		t.Fatalf("reader result %q does not record a stop", got)
	}
}

// TestAbortCaptureBoundsTheDrain witnesses the bounded-drain half of review B2 on
// EVERY host (unlike the setsid witness): with a finished channel that never
// closes, abortCapture still returns within commandWaitDelay.
func TestAbortCaptureBoundsTheDrain(t *testing.T) {
	cmd := &exec.Cmd{} // not started -> killGroup is a no-op
	never := make(chan struct{})
	done := make(chan time.Duration, 1)
	go func() {
		start := time.Now()
		abortCapture(cmd, io.NopCloser(strings.NewReader("")), io.NopCloser(strings.NewReader("")), never)
		done <- time.Since(start)
	}()
	select {
	case elapsed := <-done:
		if elapsed < commandWaitDelay-100*time.Millisecond {
			t.Fatalf("abortCapture returned early (%v); want ~%v", elapsed, commandWaitDelay)
		}
	case <-time.After(commandWaitDelay + 3*time.Second):
		t.Fatal("abortCapture did not bound the drain")
	}
}

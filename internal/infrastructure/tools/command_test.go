package tools

import (
	"context"
	"strings"
	"testing"
	"time"
)

// T034/T035 — the command tool and the reader FR-018 timeout-result path.

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
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	got, err := executeCommand{}.Execute(ctx, `{"command":"sleep 60","reason":"r"}`, testBudget)
	if err != nil {
		t.Fatalf("an observed deadline must be a nil-error result, got %v", err)
	}
	if !strings.Contains(got, "stopped at the time limit") {
		t.Fatalf("result %q does not record a stop", got)
	}
}

func TestReaderObservedDeadlineIsNilErrorTimeoutResult(t *testing.T) {
	// A deadline already in the past (no sleep needed — deterministic).
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

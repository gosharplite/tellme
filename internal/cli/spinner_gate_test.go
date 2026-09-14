package cli

import (
	"io"
	"testing"
)

// T015 [UNIT] — the spinner gate resolution (round-019 FR-006/FR-008): drawn only
// on a non-TUI prompt surface, when -r is off, and when the diagnostic stream
// (stderr) is a terminal.

func TestSpinnerGate(t *testing.T) {
	tests := []struct {
		name   string
		chrome bool
		raw    bool
		tty    bool
		want   bool
	}{
		{"prompt surface on a terminal", true, false, true, true},
		{"raw flag suppresses", true, true, true, false},
		{"non-terminal stderr suppresses", true, false, false, false},
		{"non-prompt surface suppresses", false, false, true, false},
		{"non-prompt and non-terminal", false, true, false, false},
	}
	for _, tt := range tests {
		if got := spinnerGate(turnOptions{chrome: tt.chrome, raw: tt.raw}, tt.tty); got != tt.want {
			t.Errorf("%s: spinnerGate(chrome=%v raw=%v tty=%v) = %v, want %v", tt.name, tt.chrome, tt.raw, tt.tty, got, tt.want)
		}
	}
}

func TestStderrTerminalDetectorForcedSeam(t *testing.T) {
	t.Setenv("TELL_ME_FORCE_STDERR_TTY", "1")
	if probe := stderrTerminalDetector(); !probe(nil) {
		t.Error("forced stderr probe did not report a terminal")
	}

	t.Setenv("TELL_ME_FORCE_STDERR_TTY", "")
	probe := stderrTerminalDetector()
	if probe("not-a-file") {
		t.Error("real stderr probe reported a terminal for a non-file stream")
	}
}

func TestRuntimeEnvStderrIsTerminal(t *testing.T) {
	// Falls back to the shared probe when no stderr probe is supplied.
	e := runtimeEnv{stderr: io.Discard, isTTY: func(any) bool { return true }}
	if !e.stderrIsTerminal() {
		t.Error("stderrIsTerminal did not fall back to the shared probe")
	}
	// Prefers the dedicated stderr probe when supplied.
	e2 := runtimeEnv{stderr: io.Discard, isTTY: func(any) bool { return false }, stderrTTY: func(any) bool { return true }}
	if !e2.stderrIsTerminal() {
		t.Error("stderrIsTerminal did not use the dedicated stderr probe")
	}
	// No probes at all → not a terminal.
	if (runtimeEnv{stderr: io.Discard}).stderrIsTerminal() {
		t.Error("stderrIsTerminal with no probe reported a terminal")
	}
}

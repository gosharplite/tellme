package cli

import (
	"io"
	"testing"
)

// Round-025 unit test (T007): the columns seam + the TELL_ME_FORCE_STDERR_COLS
// override that drives the spinner's row-aware clear.

func TestStderrColumnsForcedSeam(t *testing.T) {
	t.Setenv("TELL_ME_FORCE_STDERR_COLS", "40")
	if got := stderrColumns(runtimeEnv{stderr: io.Discard})(); got != 40 {
		t.Errorf("stderrColumns() = %d, want 40", got)
	}

	t.Setenv("TELL_ME_FORCE_STDERR_COLS", "not-a-number")
	if got := stderrColumns(runtimeEnv{stderr: io.Discard})(); got != 0 {
		t.Errorf("stderrColumns() with a bad override = %d, want 0 (unknown)", got)
	}

	t.Setenv("TELL_ME_FORCE_STDERR_COLS", "")
	// io.Discard is not an *os.File → the real probe returns 0 (unknown), so the
	// presenter degrades to a single-row best-effort clear.
	if got := stderrColumns(runtimeEnv{stderr: io.Discard})(); got != 0 {
		t.Errorf("stderrColumns() with a non-file stream = %d, want 0", got)
	}
}

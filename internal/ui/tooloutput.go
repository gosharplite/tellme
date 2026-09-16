package ui

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

// Round-034 live `[Tool Output]` block (specs/truth/features/cli/chat/watching-the-tool-loop.feature;
// ADR 0005 D7 / FR-010): the header, a fixed separator literal, each complete
// output line, and a closing separator. The reference literals are pinned
// verbatim (Q3): the header is `Executing... (Output shown below)` and the
// separator is 60 ASCII hyphens. A trailing partial line is DROPPED, never
// flushed (FR-010). The writer owns its own mutex, because the child's stdout and
// stderr each feed it from a separate io.Copy goroutine (ADR 0005 D7).

// ToolOutputSeparator is the reference's fixed separator literal (60 hyphens).
const ToolOutputSeparator = "------------------------------------------------------------"

// FormatToolOutputHeader renders the block header line (FR-010):
// `[HH:MM:SS] [Tool Output] Executing... (Output shown below)`.
func FormatToolOutputHeader(t time.Time) string {
	return fmt.Sprintf("[%s] [Tool Output] Executing... (Output shown below)", formatClock(t))
}

// FormatToolOutputLine renders one streamed output line (FR-010):
// `[HH:MM:SS] [Tool Output] <line>`.
func FormatToolOutputLine(t time.Time, line string) string {
	return fmt.Sprintf("[%s] [Tool Output] %s", formatClock(t), line)
}

// ToolOutputWriter streams a shell command's complete output lines as a
// `[Tool Output]` block on the diagnostic stream (FR-010). It is an io.Writer so
// the command's bounded stdout/stderr pipes can each tee into it; the mutex
// serialises those two copy goroutines, and the partial-line buffer carries a
// line split across Write calls until its terminating newline arrives.
type ToolOutputWriter struct {
	// W is the diagnostic stream the block is written to; nil disables the block.
	W io.Writer
	// Now is the injected clock seam; nil falls back to time.Now.
	Now func() time.Time

	mu   sync.Mutex
	buf  []byte
	open bool
}

// now returns the clock reading (falling back to time.Now).
func (w *ToolOutputWriter) now() time.Time {
	if w.Now != nil {
		return w.Now()
	}
	return time.Now()
}

// Begin writes the header and the opening separator (FR-010). It is a no-op on a
// nil writer or a writer already open.
func (w *ToolOutputWriter) Begin() {
	if w.W == nil {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.open {
		return
	}
	w.open = true
	w.buf = w.buf[:0]
	_, _ = fmt.Fprintln(w.W, FormatToolOutputHeader(w.now()))
	_, _ = fmt.Fprintln(w.W, ToolOutputSeparator)
}

// Write assembles complete lines and emits each as a `[Tool Output]` line; a
// trailing partial line is kept for the next call (or dropped at End). It
// consumes all input and never errors (best-effort diagnostics).
func (w *ToolOutputWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.W == nil || !w.open {
		return len(p), nil
	}
	w.buf = append(w.buf, p...)
	for {
		i := bytes.IndexByte(w.buf, '\n')
		if i < 0 {
			break
		}
		line := strings.TrimRight(string(w.buf[:i]), "\r")
		w.buf = w.buf[i+1:]
		_, _ = fmt.Fprintln(w.W, FormatToolOutputLine(w.now(), line))
	}
	return len(p), nil
}

// End drops the trailing partial line (never flushed) and writes the closing
// separator (FR-010). It is idempotent.
func (w *ToolOutputWriter) End() {
	if w.W == nil {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.open {
		return
	}
	w.open = false
	w.buf = nil // drop the trailing partial line — never flushed
	_, _ = fmt.Fprintln(w.W, ToolOutputSeparator)
}

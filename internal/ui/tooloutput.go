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
//
// Round 040 (ADR 0009 D3/D4 — supersedes ADR 0005 D7): the writer is ALSO the
// block's sole lock/state owner for the WS-A idle-gap liveness. It exposes three
// lock-scoped entry points — WriteWith (the line path, with a per-line clear
// hook), EndWith (the close-path clear hook), and withLock (the bookkeeping/idle
// query) — and stamps `lastLine` INSIDE the same critical section that emits a
// line. The coordinator admits the resume under this mutex and clears/joins before
// a line via these entry points, so the invariant is mutual exclusion + join (not
// "single writer") with lock order block-writer mutex → spinner mutex — and the
// coordinator never reaches into `mu`.

// ToolOutputSeparator is the reference's fixed separator literal (60 hyphens).
const ToolOutputSeparator = "------------------------------------------------------------"

// ToolOutputReset is the default-state (SGR reset) restore written when a block
// closes, so the terminal can never be left in a non-default state (round 038
// FR-005 — independent of FR-001's line sanitization). It is written
// UNCONDITIONALLY, including when the diagnostic stream is not a terminal (ADR
// 0007 TD-2: a no-op visually on a non-terminal, but it keeps the invariant
// independent of where the stream goes — see the ADR for the rejected
// `isatty(stderr)` gating alternative).
const ToolOutputReset = "\x1b[0m"

// FormatToolOutputHeader renders the block header line (FR-010):
// `[HH:MM:SS] [Tool Output] Executing... (Output shown below)`.
func FormatToolOutputHeader(t time.Time) string {
	return fmt.Sprintf("[%s] [Tool Output] Executing... (Output shown below)", formatClock(t))
}

// formatToolOutputHeaderColour is FormatToolOutputHeader with the round-057 grey
// accent (ADR 0027): the whole header line is wrapped grey on a colour-enabled
// terminal. The colour-off path returns the plain header verbatim.
func formatToolOutputHeaderColour(t time.Time, colour bool) string {
	return grey(FormatToolOutputHeader(t), colour)
}

// toolOutputSeparatorColour wraps the fixed separator literal grey when enabled
// (round 057 fold R-057-1: unexported — both call sites are in this file and the
// literal's owner stays internal, matching formatToolOutputHeaderColour); the
// plain path returns the pinned literal unchanged.
func toolOutputSeparatorColour(colour bool) string { return grey(ToolOutputSeparator, colour) }

// FormatToolOutputLine renders one streamed output line (FR-010):
// `[HH:MM:SS] [Tool Output] <line>`. Round 038 (FR-001, issue #78): the streamed
// content is SANITIZED — every terminal control sequence is removed — so a
// colouring/control-emitting command cannot alter the operator's terminal. This
// is presentation-only: the tool result fed to the model keeps its raw bytes.
func FormatToolOutputLine(t time.Time, line string) string {
	return fmt.Sprintf("[%s] [Tool Output] %s", formatClock(t), sanitizeControl(line))
}

// formatToolOutputLineColour is FormatToolOutputLine with the round-058 grey
// accent (ADR 0028): the WHOLE content line — the `[HH:MM:SS] [Tool Output] `
// framing plus the SANITIZED content — is wrapped grey when enabled, so the whole
// `[Tool Output]` block (header + every content line + both separators) reads as
// one grey region. The wrap is applied OUTSIDE the sanitizer, so a command's
// control bytes cannot escape it. The colour-off path returns the plain line
// verbatim (byte-identical).
func formatToolOutputLineColour(t time.Time, line string, colour bool) string {
	return grey(FormatToolOutputLine(t, line), colour)
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
	// Colour, when true, wraps EVERY `[Tool Output]` line grey — the header, each
	// streamed content line (round 058; ADR 0028), and both horizontal separators
	// (round 057; ADR 0027). The content wrap sits OUTSIDE the round-038 sanitizer,
	// so grey is the line's only escape. The plain path is byte-identical to the
	// pre-057 literals.
	Colour bool

	mu   sync.Mutex
	buf  []byte
	open bool
	// lastLine is when the block last emitted an output line (round 040, ADR 0009
	// D3/D4): seeded at Begin and stamped in the same critical section that emits
	// each line, so the WS-A idle query (withLock) measures the true quiet gap.
	lastLine time.Time
}

// now returns the clock reading (falling back to time.Now).
func (w *ToolOutputWriter) now() time.Time {
	if w.Now != nil {
		return w.Now()
	}
	return time.Now()
}

// Begin writes the header and the opening separator (FR-010). It is a no-op on a
// nil writer or a writer already open. Round 040 (ADR 0009 D3): it ALSO seeds the
// idle clock (`lastLine`), so a command that prints nothing at all still resumes
// the indicator after the idle gap (the zero-output case).
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
	now := w.now()
	w.lastLine = now
	_, _ = fmt.Fprintln(w.W, formatToolOutputHeaderColour(now, w.Colour))
	_, _ = fmt.Fprintln(w.W, toolOutputSeparatorColour(w.Colour))
}

// Write assembles complete lines and emits each as a `[Tool Output]` line; a
// trailing partial line is kept for the next call (or dropped at End). It is
// WriteWith with no per-line hook.
func (w *ToolOutputWriter) Write(p []byte) (int, error) { return w.WriteWith(p, nil) }

// WriteWith is the round-040 line-path entry point (ADR 0009 D4): for each
// COMPLETE output line it first calls beforeLine (the coordinator's synchronous
// clear) and then writes the line and stamps the idle clock — all under the
// writer's mutex, so a line can never be interleaved by a spinner frame and the
// clear is atomic with respect to the line. A nil beforeLine is a plain write. It
// consumes all input and never errors (best-effort diagnostics). The line
// splitting stays inside this method over the private `buf` (R-2).
func (w *ToolOutputWriter) WriteWith(p []byte, beforeLine func()) (int, error) {
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
		if beforeLine != nil {
			beforeLine()
		}
		now := w.now()
		_, _ = fmt.Fprintln(w.W, formatToolOutputLineColour(now, line, w.Colour))
		w.lastLine = now
	}
	return len(p), nil
}

// End drops the trailing partial line (never flushed) and writes the closing
// separator (FR-010). It is EndWith with no hook. It is idempotent. Round 038
// (FR-005): it first restores the terminal's default state (a reset), so a
// dropped trailing-partial-line reset or a command killed mid-output cannot strand
// the terminal in a non-default state.
func (w *ToolOutputWriter) End() { w.EndWith(nil) }

// EndWith is the round-040 close-path entry point (ADR 0009 D4, R-8): it calls
// beforeSeparator (the coordinator's synchronous clear) INSIDE the writer's
// critical section before the closing separator, so the clear is atomic with
// respect to any in-flight line write (a drain goroutine can still be inside
// Write at End on the trim/timeout paths). A nil hook is a plain close.
func (w *ToolOutputWriter) EndWith(beforeSeparator func()) {
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
	if beforeSeparator != nil {
		beforeSeparator()
	}
	_, _ = fmt.Fprint(w.W, ToolOutputReset)
	_, _ = fmt.Fprintln(w.W, toolOutputSeparatorColour(w.Colour))
}

// withLock is the round-040 bookkeeping/idle entry point (ADR 0009 D4, R-11): it
// runs fn under the writer's mutex, handing it the current idle duration since the
// last emitted line. The writer computes the idle under its own lock, so the
// coordinator's check + admit is one critical section (a self-locking IdleSince
// called inside would deadlock; an unlocked one would race). It is a no-op when
// the block is not open.
func (w *ToolOutputWriter) withLock(fn func(idle time.Duration)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.open {
		return
	}
	fn(w.now().Sub(w.lastLine))
}

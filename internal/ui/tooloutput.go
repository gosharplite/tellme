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

// FormatToolOutputLine renders one streamed output line (FR-010):
// `[HH:MM:SS] [Tool Output] <line>`. Round 038 (FR-001, issue #78): the streamed
// content is SANITIZED — every terminal control sequence is removed — so a
// colouring/control-emitting command cannot alter the operator's terminal. This
// is presentation-only: the tool result fed to the model keeps its raw bytes.
func FormatToolOutputLine(t time.Time, line string) string {
	return fmt.Sprintf("[%s] [Tool Output] %s", formatClock(t), sanitizeControl(line))
}

// sanitizeControl removes terminal control data from a streamed output line
// (round 038 FR-001; ADR 0007). The removed class is precisely:
//
//   - every **7-bit** ANSI escape sequence introduced by ESC (0x1b): CSI
//     (ESC `[` … final), OSC (ESC `]` … BEL | ESC `\`), and a generic ESC + optional
//     intermediates (0x20–0x2F) + one **ASCII** final byte;
//   - every stray C0 control byte (0x00–0x1F) **except TAB** (0x09) — this
//     includes interior CR (0x0D), so in-place `\r` progress output concatenates;
//   - DEL (0x7f).
//
// Out of scope (untouched): the 8-bit C1 control range (0x80–0x9F) and all other
// bytes ≥ 0x80. The sanitizer only REMOVES bytes; it never INTRODUCES invalid
// UTF-8 — the ESC consumption is ASCII-gated (so it can never split a multi-byte
// rune), while the bytes it forwards are passed through verbatim (so a genuinely
// binary source can still arrive as invalid UTF-8 and be forwarded unchanged).
//
// The ESC-scan window is bounded PER KIND (csiScanLimit / oscScanLimit): a
// sequence whose terminator lies within its window is removed in FULL, however
// long (an OSC-8 hyperlink's URL is legitimately hundreds of bytes), while a
// sequence whose terminator lies beyond the window — or an unterminated one —
// drops only the ESC and continues, so a mangled/binary fragment cannot swallow a
// line's visible text. Pure, and allocation-free on a control-free line.
func sanitizeControl(s string) string {
	if !strings.ContainsFunc(s, isControlRune) {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); {
		switch c := s[i]; {
		case c == 0x1b: // ESC — consume the whole (7-bit) sequence
			i += escSequenceLen(s[i:])
		case (c < 0x20 && c != '\t') || c == 0x7f:
			i++ // drop a stray control byte
		default:
			b.WriteByte(c)
			i++
		}
	}
	return b.String()
}

// The per-kind bounds on the ESC-scan window (ADR 0007 D2). CSI parameters are
// numeric/intermediate and short by nature; OSC carries a string — a window title
// or an OSC-8 hyperlink URL — which is legitimately long, so its window is much
// wider. Either way the window only bounds an UNTERMINATED (or terminator-beyond-
// window) sequence so a blob cannot swallow the line; a terminated sequence inside
// its window is always removed in full.
const (
	csiScanLimit = 128
	oscScanLimit = 1024
)

// isControlRune reports whether r is terminal control data (ESC, a C0 control
// other than TAB, or DEL).
func isControlRune(r rune) bool {
	return r == 0x1b || (r < 0x20 && r != '\t') || r == 0x7f
}

// escSequenceLen returns the byte length of the 7-bit escape sequence starting
// at s[0] == ESC (0x1b). It consumes ASCII only, so it never splits a multi-byte
// rune.
func escSequenceLen(s string) int {
	if len(s) < 2 {
		return 1
	}
	switch s[1] {
	case '[':
		return csiLen(s)
	case ']':
		return oscLen(s)
	default:
		return genericEscLen(s)
	}
}

// csiLen returns the length of a CSI (ESC `[`) sequence: parameter/intermediate
// bytes then a final byte 0x40–0x7E, within the CSI window. A sequence whose
// terminator lies beyond the window (or is absent) drops only the ESC (returns 1)
// so it cannot swallow the line's remaining text.
func csiLen(s string) int {
	limit := min(len(s), csiScanLimit)
	for i := 2; i < limit; i++ {
		if s[i] >= 0x40 && s[i] <= 0x7e {
			return i + 1
		}
	}
	return 1
}

// oscLen returns the length of an OSC (ESC `]`) sequence: terminated by BEL
// (0x07) or ST (ESC `\`), within the OSC window. A sequence whose terminator lies
// beyond the window (or is absent) drops only the ESC (returns 1).
func oscLen(s string) int {
	limit := min(len(s), oscScanLimit)
	for i := 2; i < limit; i++ {
		if s[i] == 0x07 {
			return i + 1
		}
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '\\' {
			return i + 2
		}
	}
	return 1
}

// genericEscLen returns the length of a generic ESC sequence: ESC + optional
// intermediates (0x20–0x2F) + one **ASCII** final byte (ADR 0007 B1: the final
// byte is ASCII-gated, so ESC + a multi-byte rune drops the ESC and keeps the
// rune intact instead of decapitating it into invalid UTF-8). The intermediate
// run is itself bounded so a long one cannot swallow the line.
func genericEscLen(s string) int {
	i := 1
	for i < len(s) && i < csiScanLimit && s[i] >= 0x20 && s[i] <= 0x2f {
		i++
	}
	if i < len(s) && s[i] < 0x80 {
		i++
	}
	return i
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
// separator (FR-010). It is idempotent. Round 038 (FR-005): it first restores the
// terminal's default state (a reset), so a dropped trailing-partial-line reset or
// a command killed mid-output cannot strand the terminal in a non-default state.
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
	_, _ = fmt.Fprint(w.W, ToolOutputReset)
	_, _ = fmt.Fprintln(w.W, ToolOutputSeparator)
}

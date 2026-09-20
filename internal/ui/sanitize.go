package ui

import "strings"

// Terminal-safe line policy (round 039, issue #80; ADR 0008, superseding ADR
// 0007). This file owns the ONE definition of the terminal-control class that is
// removed from a `[Tool …]` diagnostic line. It is applied by EVERY `[Tool …]`
// formatter that renders model-authored or externally-sourced text — the
// `[Tool Output]` content lines (`tooloutput.go`), and the `[Tool Reason]` /
// `[Tool Result]` / `[Tool Action]` / `[Tool Warning]` lines (`toolcall.go`) — so the class is
// defined once and cannot drift between siblings. It is a pure presentation-seam
// rule: the tool **result** fed to the model and the call's raw/stored arguments
// are separate consumers and keep their raw bytes.

// sanitizeControl removes terminal control data from a rendered line (round 038
// FR-001; ADR 0007 D2, scope generalized by round 039 ADR 0008). The removed
// class is precisely:
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
// run is bounded by the CSI window (a shared constant — the final byte is a single
// ASCII byte, so the run is the only unbounded part) so a long one cannot swallow
// the line.
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

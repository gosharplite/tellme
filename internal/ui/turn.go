package ui

import (
	"fmt"
	"strings"
	"time"
)

// Round-017 operator turn chrome (specs/truth/features/cli/chat/presenting-the-turn.feature).
// A prompt-bearing turn on the non-TUI surfaces opens like tell-me-go: an
// input-capture acknowledgement, a blank line, an 80-column horizontal rule, and
// a `╭─⠿ Turn <N> - <mode>` header — all on the diagnostic stream, in plain text
// (round-017 Decision 3: no ANSI this round). The pre-flight payload line
// (FormatPayloadStatus) is emitted inside the frame, unchanged.

// turnRuleWidth is the reference's fixed horizontal-rule width (80 `─`).
const turnRuleWidth = 80

// turnGlyph is the reference's turn-header spinner glyph.
const turnGlyph = "⠿"

// FormatInputCaptured renders the reference's input-capture acknowledgement line
// (no trailing newline): `[HH:MM:SS] Input captured. Processing...`. The
// timestamp comes from the caller's injected clock seam.
func FormatInputCaptured(t time.Time) string {
	return fmt.Sprintf("[%s] Input captured. Processing...", formatClock(t))
}

// turnRule is the reference's fixed-width horizontal rule, computed once at
// package init (the `─` rune is a 3-byte UTF-8 sequence, so repeating it per
// call would allocate — round-017 implementation-review finding 2).
var turnRule = strings.Repeat("─", turnRuleWidth)

// TurnRule returns the reference's fixed-width horizontal rule (80 `─`).
func TurnRule() string { return turnRule }

// FormatTurnHeader renders the reference's turn header: `╭─⠿ Turn <N> - <mode>`.
// The ` - <mode>` suffix is omitted when mode is empty.
func FormatTurnHeader(turn int, mode string) string {
	suffix := ""
	if mode != "" {
		suffix = " - " + mode
	}
	return fmt.Sprintf("╭─%s Turn %d%s", turnGlyph, turn, suffix)
}

// FormatTurnOpening renders the frame's opening block, each line newline-
// terminated: a leading blank line, the horizontal rule, and the turn header.
// The caller writes it before the pre-flight payload line, then a trailing blank
// line after it (FormatTurnGap).
func FormatTurnOpening(turn int, mode string) string {
	return "\n" + TurnRule() + "\n" + FormatTurnHeader(turn, mode) + "\n"
}

// FormatTurnGap is the blank line that separates the frame from the answer.
func FormatTurnGap() string { return "\n" }

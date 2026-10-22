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
// (FormatPayloadEstimate) is emitted inside the frame.

// turnRuleWidth is the reference's fixed horizontal-rule width (80 `─`).
const turnRuleWidth = 80

// turnGlyph is the reference's turn-header spinner glyph.
const turnGlyph = "⠿"

// FormatInputCaptured renders the reference's input-capture acknowledgement line
// (no trailing newline): `[HH:MM:SS] Input captured. Processing...`. The
// timestamp comes from the caller's injected clock seam.
//
// Round 097 (ADR 0067): this stays the SINGLE plain entry point — the colour-off
// path is byte-identical to it (the round-046 no-orphan rule), and the accent
// lives in formatInputCapturedColour below.
func FormatInputCaptured(t time.Time) string {
	return fmt.Sprintf("[%s] Input captured. Processing...", formatClock(t))
}

// formatInputCapturedColour is FormatInputCaptured with the round-097 blue accent
// (ADR 0067 — the FIFTH chrome element): the WHOLE line, its timestamp included,
// is wrapped in the reference's bright-blue SGR pair, exactly as
// formatToolReasonColour wraps the whole `[Tool Reason]` line (round 054).
//
// The colour-off path returns FormatInputCaptured verbatim, so the round-054 gate
// (terminal stderr + `-r` off) leaves the ack byte-identical to the plain form.
// Blue (not green) because green's dominant instance is the `$` session cost on
// the `Ready` line three lines below, while blue's only other meaning — the `-l`
// listing's `[USER]` header — is the same operator-side concept on a DIFFERENT
// surface (stdout), so the two never share a block.
func formatInputCapturedColour(t time.Time, colour bool) string {
	plain := FormatInputCaptured(t)
	if !colour {
		return plain
	}
	return blue(plain, true)
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

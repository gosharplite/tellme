package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/gosharplite/tellme/internal/domain/render"
)

// T010 [UNIT] — the turn-chrome formatter (deterministic formatting).

func TestFormatInputCaptured(t *testing.T) {
	got := FormatInputCaptured(time.Date(2026, 9, 14, 10, 20, 0, 0, time.UTC))
	want := "[10:20:00] Input captured. Processing..."
	if got != want {
		t.Fatalf("FormatInputCaptured = %q, want %q", got, want)
	}
}

func TestTurnRuleIsEightyColumns(t *testing.T) {
	rule := TurnRule()
	if n := len([]rune(rule)); n != 80 {
		t.Fatalf("rule width = %d, want 80", n)
	}
	if strings.Trim(rule, "─") != "" {
		t.Fatalf("rule carries non-rule bytes: %q", rule)
	}
}

func TestFormatTurnHeader(t *testing.T) {
	if got, want := FormatTurnHeader(1, "butler"), "╭─⠿ Turn 1 - butler"; got != want {
		t.Fatalf("FormatTurnHeader = %q, want %q", got, want)
	}
	if got, want := FormatTurnHeader(3, ""), "╭─⠿ Turn 3"; got != want {
		t.Fatalf("FormatTurnHeader(no mode) = %q, want %q", got, want)
	}
}

func TestFormatTurnOpening(t *testing.T) {
	got := FormatTurnOpening(2, "butler")
	lines := strings.Split(strings.TrimSuffix(got, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("opening has %d lines, want 3; got %q", len(lines), got)
	}
	if lines[0] != "" {
		t.Fatalf("line 0 = %q, want the leading blank line", lines[0])
	}
	if lines[1] != TurnRule() {
		t.Fatalf("line 1 = %q, want the 80-column rule", lines[1])
	}
	if lines[2] != "╭─⠿ Turn 2 - butler" {
		t.Fatalf("line 2 = %q, want the turn header", lines[2])
	}
	if !strings.HasSuffix(got, "\n") {
		t.Fatalf("opening is not newline-terminated: %q", got)
	}
}

func TestFormatTurnGapIsBlankLine(t *testing.T) {
	if got := FormatTurnGap(); got != "\n" {
		t.Fatalf("FormatTurnGap = %q, want a single blank line", got)
	}
}

// The round-017 chrome's inner line is the round-057 estimated payload line
// (`+<delta> ~<n>`; ADR 0027); the measured line keeps the round-009 shape.
func TestPayloadEstimateLineText(t *testing.T) {
	got := FormatPayloadEstimate(render.PayloadStatus{At: time.Date(2026, 9, 14, 13, 16, 25, 0, time.UTC), Tokens: 1234, Mode: "butler", Model: "deepseek-v4-flash"})
	want := "[13:16:25] Payload: +0 ~1234 tokens - butler - deepseek-v4-flash"
	if got != want {
		t.Fatalf("estimated payload line changed: got %q, want %q", got, want)
	}
}

// Round 097 (ADR 0067) — the input-capture acknowledgement's blue accent is the
// FIFTH chrome element. These pins carry the parts the E2E cannot see cheaply:
// the byte-identity of the colour-off path, and the whole-line (timestamp
// included) scope of the wrap.

func TestFormatInputCapturedColour_OffIsByteIdentical(t *testing.T) {
	ts := time.Date(2026, 10, 22, 13, 2, 41, 0, time.UTC)
	if got, want := formatInputCapturedColour(ts, false), FormatInputCaptured(ts); got != want {
		t.Fatalf("colour-off path must be byte-identical to the plain form: got %q, want %q", got, want)
	}
}

func TestFormatInputCapturedColour_OnWrapsTheWholeLine(t *testing.T) {
	ts := time.Date(2026, 10, 22, 13, 2, 41, 0, time.UTC)
	got := formatInputCapturedColour(ts, true)
	// The bright-blue pair encloses the TIMESTAMP as well as the message — the
	// whole-line unit (ADR 0067 D1), matching the `[Tool Reason]` precedent.
	want := colorBlue + FormatInputCaptured(ts) + colorReset
	if got != want {
		t.Fatalf("colour-on wrap = %q, want %q", got, want)
	}
	if !strings.HasPrefix(got, "\x1b[1;34m") {
		t.Fatalf("the accent must be the bright blue SGR pair: %q", got)
	}
	// The visible text is unchanged by the wrap (the pair is the only escape).
	if stripped := strings.TrimSuffix(strings.TrimPrefix(got, colorBlue), colorReset); stripped != FormatInputCaptured(ts) {
		t.Fatalf("the wrap altered the visible text: %q", stripped)
	}
}

// The `Lines` adapter (the production render.Lines implementation) must honour
// its colour flag on the acknowledgement — the wiring the CLI gate relies on.
func TestLines_InputCapturedHonoursTheColourFlag(t *testing.T) {
	ts := time.Date(2026, 10, 22, 13, 2, 41, 0, time.UTC)
	if got, want := NewLines(false).InputCaptured(ts), FormatInputCaptured(ts); got != want {
		t.Fatalf("Lines(false).InputCaptured = %q, want the plain form %q", got, want)
	}
	if got, want := NewLines(true).InputCaptured(ts), colorBlue+FormatInputCaptured(ts)+colorReset; got != want {
		t.Fatalf("Lines(true).InputCaptured = %q, want %q", got, want)
	}
}

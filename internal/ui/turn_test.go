package ui

import (
	"strings"
	"testing"
	"time"
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

// The round-009 payload line text is the round-017 chrome's inner line and MUST
// stay unchanged (research Decision 3).
func TestPayloadStatusTextUnchanged(t *testing.T) {
	got := FormatPayloadStatus(time.Date(2026, 9, 14, 13, 16, 25, 0, time.UTC), 1234, 1000000, "butler", "deepseek-v4-flash", true)
	want := "[13:16:25] Payload: ~1234/1000000 tokens - butler - deepseek-v4-flash"
	if got != want {
		t.Fatalf("payload line changed: got %q, want %q", got, want)
	}
}

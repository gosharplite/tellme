package ui

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// Round 036 (issue #74): the model-authored `reason` is the ONLY free-text field
// in the tool log, and FormatToolReason must sanitize + cap it like its siblings
// — fold embedded newlines/carriage returns to spaces, trim the surrounding
// whitespace, and cap the folded value at reasonValueCap runes (one U+2026 inside
// the cap, rune-safe). These pins use HOSTILE fixtures (the round-022 B1 lesson:
// the original regression was untested because the fixtures were single-line).

func TestFormatToolReason_FoldsNewline(t *testing.T) {
	if got, want := FormatToolReason(r034Clock, "a\nb"), "[20:29:51] [Tool Reason] a b"; got != want {
		t.Errorf("FormatToolReason newline fold = %q; want %q", got, want)
	}
	if strings.ContainsAny(FormatToolReason(r034Clock, "a\nb"), "\n\r") {
		t.Errorf("the folded reason still carries a control newline")
	}
}

func TestFormatToolReason_FoldsCarriageReturn(t *testing.T) {
	got := FormatToolReason(r034Clock, "a\rb")
	const want = "[20:29:51] [Tool Reason] a b"
	if got != want {
		t.Errorf("FormatToolReason carriage-return fold = %q; want %q (a \\r must not truncate the row)", got, want)
	}
	if strings.Contains(got, "\r") {
		t.Errorf("the folded reason still carries \\r")
	}
}

func TestFormatToolReason_TrimsSurroundingWhitespace(t *testing.T) {
	if got, want := FormatToolReason(r034Clock, "  spaced  "), "[20:29:51] [Tool Reason] spaced"; got != want {
		t.Errorf("FormatToolReason trim = %q; want %q", got, want)
	}
}

func TestFormatToolReason_CapsAndFoldsCombined(t *testing.T) {
	if got, want := FormatToolReason(r034Clock, "a\r\nb"), "[20:29:51] [Tool Reason] a  b"; got != want {
		t.Errorf("FormatToolReason combined fold = %q; want %q", got, want)
	}
}

// wantReasonCap is the reason cap pinned by the truth (`chat/dsl.md`, round 036 :
// reasonValueCap = 200) and by research Decision 2. It is named here so the pin
// compiles RED-first (the product constant lands in Phase 4).
const wantReasonCap = 200

func TestFormatToolReason_ValuesAreRuneCapped(t *testing.T) {
	long := strings.Repeat("x", 201) // exceeds the 200-rune reason cap
	got := FormatToolReason(r034Clock, long)
	const prefix = "[20:29:51] [Tool Reason] "
	value := strings.TrimPrefix(got, prefix)
	if n := utf8.RuneCountInString(value); n != wantReasonCap {
		t.Errorf("capped reason is %d runes; want %d", n, wantReasonCap)
	}
	if !strings.HasSuffix(value, "…") {
		t.Errorf("capped reason does not end with one U+2026; value=%q", value)
	}
	if strings.Contains(value, "..") {
		t.Errorf("capped reason used ASCII dots; value=%q", value)
	}
}

func TestFormatToolReason_CapCutOnRuneBoundary(t *testing.T) {
	// A multi-byte rune must never be split by the cap (rune-safe cut).
	long := strings.Repeat("é", 300) // 300 runes, 600 bytes
	got := FormatToolReason(r034Clock, long)
	value := strings.TrimPrefix(got, "[20:29:51] [Tool Reason] ")
	if n := utf8.RuneCountInString(value); n != wantReasonCap {
		t.Errorf("capped multi-byte reason is %d runes; want %d", n, wantReasonCap)
	}
	if !utf8.ValidString(value) {
		t.Errorf("the cap split a multi-byte rune; value=%q", value)
	}
}

func TestFormatToolReason_CleanReasonUnchanged(t *testing.T) {
	// A well-formed single-line in-cap reason renders byte-identically to before.
	if got, want := FormatToolReason(r034Clock, "checking the launch code"), "[20:29:51] [Tool Reason] checking the launch code"; got != want {
		t.Errorf("clean reason rendering = %q; want %q", got, want)
	}
}

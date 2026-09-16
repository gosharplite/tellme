package ui

import (
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

// Round-034 T020 unit pins for the pure decomposed tool-call formatters (ADR 0005
// D5/D6): the rune-safe caps, the single U+2026 inside the cap, sorted keys +
// json.Number, the `reason`-exclusion, and the unparseable-args → `<tool>()` branch.

var r034Clock = time.Date(2026, 9, 16, 20, 29, 51, 0, time.UTC)

func TestFormatToolEngine(t *testing.T) {
	if got, want := FormatToolEngine(r034Clock, 2, 5), "[20:29:51] [Tool Engine] Step 2/5"; got != want {
		t.Errorf("FormatToolEngine = %q; want %q", got, want)
	}
}

func TestFormatToolReason(t *testing.T) {
	if got, want := FormatToolReason(r034Clock, "checking the launch code"), "[20:29:51] [Tool Reason] checking the launch code"; got != want {
		t.Errorf("FormatToolReason = %q; want %q", got, want)
	}
}

func TestFormatToolAction_SortedKeysReasonExcluded(t *testing.T) {
	got := FormatToolAction(r034Clock, "read_files", `{"filepaths":["notes.txt"],"reason":"checking","alpha":1}`)
	want := `[20:29:51] [Tool Action] read_files(alpha: 1, filepaths: ["notes.txt"])`
	if got != want {
		t.Errorf("FormatToolAction = %q; want %q", got, want)
	}
}

func TestFormatToolAction_NumberRendersRawLiteral(t *testing.T) {
	got := FormatToolAction(r034Clock, "t", `{"max_output_tokens":1000000}`)
	want := "[20:29:51] [Tool Action] t(max_output_tokens: 1000000)"
	if got != want {
		t.Errorf("FormatToolAction = %q; want %q (json.Number, not 1e+06)", got, want)
	}
}

func TestFormatToolAction_UnparseableRendersEmptyArgs(t *testing.T) {
	for _, in := range []string{`not-json`, `[1,2]`, ``, `{`} {
		if got, want := FormatToolAction(r034Clock, "t", in), "[20:29:51] [Tool Action] t()"; got != want {
			t.Errorf("FormatToolAction(%q) = %q; want %q", in, got, want)
		}
	}
}

func TestFormatToolAction_ValuesAreRuneCapped(t *testing.T) {
	long := strings.Repeat("a", 190) // exceeds the 189-rune cap
	got := FormatToolAction(r034Clock, "write_file", `{"content":"`+long+`"}`)
	const prefix = "[20:29:51] [Tool Action] write_file(content: "
	value := strings.TrimSuffix(strings.TrimPrefix(got, prefix), ")")
	if got2 := utf8.RuneCountInString(value); got2 != argValueCap {
		t.Errorf("capped value is %d runes; want %d (value=%q)", got2, argValueCap, value)
	}
	if !strings.HasSuffix(value, "…") {
		t.Errorf("capped value does not end with one U+2026; value=%q", value)
	}
	if strings.Contains(value, "..") {
		t.Errorf("capped value used ASCII dots; value=%q", value)
	}
}

func TestFormatToolResult_FoldsAndCaps(t *testing.T) {
	long := strings.Repeat("b", 201) // exceeds the 200-rune cap
	got := FormatToolResult(r034Clock, "read_files", long)
	const prefix = "[20:29:51] [Tool Result] read_files: "
	snippet := strings.TrimPrefix(got, prefix)
	if n := utf8.RuneCountInString(snippet); n != resultValueCap {
		t.Errorf("capped snippet is %d runes; want %d", n, resultValueCap)
	}
	if !strings.HasSuffix(snippet, "…") {
		t.Errorf("capped snippet does not end with one U+2026; snippet=%q", snippet)
	}
	if got, want := FormatToolResult(r034Clock, "t", "a\nb"), "[20:29:51] [Tool Result] t: a b"; got != want {
		t.Errorf("FormatToolResult folding = %q; want %q", got, want)
	}
}

func TestCapRunes_RuneBoundary(t *testing.T) {
	// A multi-byte rune must never be split (the cut is on a rune boundary).
	s := strings.Repeat("é", 5) // 5 runes, 10 bytes
	if got := capRunes(s, 3); utf8.RuneCountInString(got) != 3 || !strings.HasSuffix(got, "…") {
		t.Errorf("capRunes rune-boundary cut = %q", got)
	}
}

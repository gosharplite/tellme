package ui

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// Round 039 (issue #80; ADR 0008): the terminal-safe-line policy is generalized
// to EVERY `[Tool …]` formatter — a model-authored reason, a tool-result snippet,
// and an argument key/value are all control-free, like the `[Tool Output]` content
// lines. Hostile fixtures at the pure-formatter layer (no pty, no process). The
// `[Tool Output]` byte-identical regression pin is the existing round-038
// `TestFormatToolOutputLineStripsControlSequences` (unchanged, re-run after the
// sanitizer moved into `sanitize.go`).

// assertControlFreeUTF8 fails if a rendered line carries an ESC/C0/DEL byte or is
// invalid UTF-8.
func assertControlFreeUTF8(t *testing.T, got string) {
	t.Helper()
	for i := 0; i < len(got); i++ {
		c := got[i]
		if c == 0x1b || (c < 0x20 && c != '\t') || c == 0x7f {
			t.Errorf("rendered line carries a control byte (0x%02x): %q", c, got)
			return
		}
	}
	if !utf8.ValidString(got) {
		t.Errorf("rendered line is not valid UTF-8: %q", got)
	}
}

func TestFormatToolReason_StripsControlSequences(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"sgr set with reset", "check\x1b[31m failed\x1b[0m", "check failed"},
		{"sgr set with no reset", "\x1b[31mchecking", "checking"},
		{"osc title", "\x1b]0;t\x07checking", "checking"},
		{"cursor hide", "\x1b[?25lchecking", "checking"},
		{"tab preserved", "check\tin", "check\tin"},
		{"utf8 preserved", "héllo — 世界", "héllo — 世界"},
		{"esc before a multibyte rune", "a\x1b日本", "a日本"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FormatToolReason(r034Clock, tc.in)
			want := "[20:29:51] [Tool Reason] " + tc.want
			if got != want {
				t.Errorf("FormatToolReason(%q) = %q; want %q", tc.in, got, want)
			}
			assertControlFreeUTF8(t, got)
		})
	}

	// The cap still bounds the VISIBLE output: sanitize runs BEFORE the rune cap,
	// so a stripped sequence does not consume cap budget.
	long := strings.Repeat("x", 250) + "\x1b[31m"
	got := FormatToolReason(r034Clock, long)
	if n := len([]rune(strings.TrimPrefix(got, "[20:29:51] [Tool Reason] "))); n != reasonValueCap {
		t.Errorf("reason cap after sanitize = %d runes; want %d", n, reasonValueCap)
	}
}

func TestFormatToolResult_StripsControlSequences(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"sgr set with reset", "line \x1b[31mred\x1b[0m", "line red"},
		{"osc title", "\x1b]0;t\x07result", "result"},
		{"utf8 preserved", "héllo — 世界", "héllo — 世界"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FormatToolResult(r034Clock, "read_files", tc.in)
			assertControlFreeUTF8(t, got)
			if !strings.Contains(got, tc.want) {
				t.Errorf("FormatToolResult(%q) = %q; want it to contain %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestFormatToolAction_StripsControlSequencesInKeysAndValues(t *testing.T) {
	cases := []struct{ name, args string }{
		{"value carries sgr", `{"content":"\u001b[31mhello\u001b[0m"}`},
		{"key carries sgr", `{"\u001b[31mcontent":"hello"}`},
		{"value carries osc", `{"content":"\u001b]0;t\u0007hello"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FormatToolAction(r034Clock, "write_file", tc.args)
			assertControlFreeUTF8(t, got)
			if !strings.Contains(got, "hello") {
				t.Errorf("FormatToolAction(%q) = %q; want the visible value preserved", tc.args, got)
			}
			if strings.Contains(got, "reason") {
				t.Errorf("FormatToolAction(%q) = %q; the reason key must stay excluded", tc.args, got)
			}
		})
	}
}

func TestFormatToolAction_KeyOrderIsSanitizedThenSorted(t *testing.T) {
	// Round-039 review RF-2: the key list is sorted AFTER sanitizing, so a raw key
	// carrying control data cannot render out of ascending order. The raw keys are
	// "\x1b[31mz" (0x1b sorts before 'a') and "a"; sorted-after-sanitize ⇒ "a", "z".
	got := FormatToolAction(r034Clock, "write_file", `{"\u001b[31mz":1,"a":2}`)
	if !strings.Contains(got, "a: 2, z: 1") {
		t.Errorf("hostile-key order = %q; want the sanitized keys sorted ascending (a, z)", got)
	}
	assertControlFreeUTF8(t, got)
}

func TestFormatToolAction_FoldsKeyNewlineLikeAValue(t *testing.T) {
	// Round-039 review RF-2: a key folds `\n` to a space exactly like a value (one
	// input class, one rendering) — not deleted as before.
	got := FormatToolAction(r034Clock, "write_file", `{"a\u000ab":1}`)
	if !strings.Contains(got, "a b: 1") {
		t.Errorf("key fold = %q; want the key newline folded to a space (a b)", got)
	}
}

func TestToolReasonRenders(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"checking", true},
		{"   ", false},                    // whitespace-only
		{"\n", false},                     // newline-only
		{"\x1b[31m", false},               // escape-only → renders no line (round 039)
		{"\x1b]0;title\x07", false},       // osc-only
		{"\x1b[31mchecking\x1b[0m", true}, // visible text survives
	}
	for _, tc := range cases {
		if got := ToolReasonRenders(tc.in); got != tc.want {
			t.Errorf("ToolReasonRenders(%q) = %v; want %v", tc.in, got, tc.want)
		}
	}
}

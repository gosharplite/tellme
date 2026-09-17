package ui

import (
	"strings"
	"testing"
	"time"
)

// Round-038 unit pins (issues #78): the [Tool Output] content is sanitized of
// terminal control data (FR-001), and the block always closes in a neutral state
// (FR-005). Hostile fixture at the pure formatter/writer layer — no pty, no
// process (research D5).

func TestFormatToolOutputLineStripsControlSequences(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"sgr set with no reset", "\x1b[31mERROR: failed", "ERROR: failed"},
		{"sgr set with reset", "\x1b[1;31mERROR\x1b[0m: failed", "ERROR: failed"},
		{"osc window title (bel)", "\x1b]0;title\x07visible", "visible"},
		{"osc window title (st)", "\x1b]0;title\x1b\\visible", "visible"},
		{"cursor hide", "\x1b[?25lworking...", "working..."},
		{"stray bel", "ding\x07", "ding"},
		{"esc charset designation", "\x1b(Bplain", "plain"},
		{"tab preserved", "a\tb", "a\tb"},
		{"utf8 preserved", "héllo — 世界", "héllo — 世界"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FormatToolOutputLine(r034Clock, tc.in)
			want := "[20:29:51] [Tool Output] " + tc.want
			if got != want {
				t.Errorf("FormatToolOutputLine(%q) = %q; want %q", tc.in, got, want)
			}
			if strings.ContainsRune(got, 0x1b) {
				t.Errorf("content line still carries an ESC byte: %q", got)
			}
		})
	}
}

// contentLines returns the emitted `[Tool Output] <text>` content lines
// (excluding the header), so a check can be scoped off the block's own close
// restore (a deliberate ESC).
func contentLines(stream string) []string {
	var out []string
	for _, ln := range strings.Split(stream, "\n") {
		i := strings.Index(ln, "] [Tool Output] ")
		if i < 0 {
			continue
		}
		if strings.HasPrefix(ln[i+len("] [Tool Output] "):], "Executing... (Output shown below)") {
			continue
		}
		out = append(out, ln)
	}
	return out
}

func TestToolOutputWriterStripsSequenceSplitAcrossWrites(t *testing.T) {
	var sb strings.Builder
	w := &ToolOutputWriter{W: &sb, Now: func() time.Time { return r034Clock }}
	w.Begin()
	// The colour set is split across two Writes and completed by the newline; the
	// writer assembles the line before sanitizing.
	_, _ = w.Write([]byte("\x1b[3"))
	_, _ = w.Write([]byte("1mboom\n"))
	w.End()

	lines := contentLines(sb.String())
	if len(lines) != 1 {
		t.Fatalf("want exactly one content line; got %d: %q", len(lines), sb.String())
	}
	if strings.ContainsRune(lines[0], 0x1b) {
		t.Errorf("split escape survived into the content line: %q", lines[0])
	}
	if !strings.Contains(lines[0], "boom") {
		t.Errorf("visible text lost: %q", lines[0])
	}
}

func TestToolOutputWriterEndRestoresNeutralState(t *testing.T) {
	var sb strings.Builder
	w := &ToolOutputWriter{W: &sb, Now: func() time.Time { return r034Clock }}
	w.Begin()
	_, _ = w.Write([]byte("\x1b[31mx\x1b[0m\n"))
	w.End()

	got := sb.String()
	lastSep := strings.LastIndex(got, ToolOutputSeparator)
	if lastSep < 0 {
		t.Fatalf("no closing separator: %q", got)
	}
	reset := strings.LastIndex(got, ToolOutputReset)
	if reset < 0 || reset > lastSep {
		t.Errorf("no neutral restore before the closing separator: %q", got)
	}
}

func TestToolOutputWriterStartFailureRestoresNeutralState(t *testing.T) {
	// The command start-failure shape: Begin then End with no output still closes
	// neutral.
	var sb strings.Builder
	w := &ToolOutputWriter{W: &sb, Now: func() time.Time { return r034Clock }}
	w.Begin()
	w.End()
	if !strings.Contains(sb.String(), ToolOutputReset) {
		t.Errorf("start-failure close emitted no restore: %q", sb.String())
	}
}

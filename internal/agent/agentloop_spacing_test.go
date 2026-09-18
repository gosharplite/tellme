package agent

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/gosharplite/tellme/internal/domain/llm"
)

// Round 039 (operator spacing request; ADR 0008): the loop's begin block for EACH
// tool call is preceded by exactly one blank line — per call, so a k-call round
// emits k blanks. The blank is tied to the call BLOCK, not the reason line, so a
// reason-less call's action line is also preceded by a blank.
//
// Round 046 (R4 of #92, ADR 0015): the loop owns the SCHEDULE (this blank + the
// line order); the fake renderer supplies marker lines, so the schedule is pinned
// without importing internal/ui.

// logLines splits a captured log into lines.
func logLines(log string) []string { return strings.Split(log, "\n") }

// lineBeforeContains reports whether the line immediately before the first line
// containing marker equals want (want == "" means a blank line precedes it).
func lineBeforeContains(log, marker, want string) bool {
	lines := logLines(log)
	for i, l := range lines {
		if strings.Contains(l, marker) {
			if i == 0 {
				return false
			}
			return lines[i-1] == want
		}
	}
	return false
}

// blockStartsPrecededByBlank counts the tool-call begin blocks and reports how
// many are preceded by a blank line. A block's start is its reason line when the
// line before the action line is a reason line, else the action line itself.
func blockStartsPrecededByBlank(log string) (total, blank int) {
	lines := logLines(log)
	for i, l := range lines {
		if !strings.Contains(l, "ACTION ") {
			continue
		}
		total++
		begin := i
		if i > 0 && strings.Contains(lines[i-1], "REASON ") {
			begin = i - 1
		}
		if begin > 0 && lines[begin-1] == "" {
			blank++
		}
	}
	return total, blank
}

func TestLogSeparatesEachCallBeginBlockWithABlankLine(t *testing.T) {
	var buf bytes.Buffer
	// One round, TWO tool calls, each with its own reason.
	a := newLogLoop(&buf,
		llm.ToolCall{ID: "c1", Name: "read_files", Arguments: `{"filepaths":["a.txt"],"reason":"first"}`},
		llm.ToolCall{ID: "c2", Name: "read_files", Arguments: `{"filepaths":["b.txt"],"reason":"second"}`},
	)
	if _, err := a.Run(context.Background(), "q", nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
	log := buf.String()
	total, blank := blockStartsPrecededByBlank(log)
	if total != 2 {
		t.Fatalf("expected 2 tool-call begin blocks; got %d (log=%q)", total, log)
	}
	if blank != 2 {
		t.Errorf("expected both call begin blocks preceded by a blank line; got %d/2 (log=%q)", blank, log)
	}
}

func TestLogBeginBlockLeadsWithReasonThenAction(t *testing.T) {
	var buf bytes.Buffer
	a := newLogLoop(&buf, llm.ToolCall{ID: "c1", Name: "read_files", Arguments: `{"filepaths":["a.txt"],"reason":"because"}`})
	if _, err := a.Run(context.Background(), "q", nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
	log := buf.String()
	// The blank precedes the reason line (the block's first line), which is then
	// immediately followed by the action line (no blank between reason and action).
	if !lineBeforeContains(log, "REASON because", "") {
		t.Errorf("the reason line was not preceded by a blank line; log=%q", log)
	}
	if !lineBeforeContains(log, "ACTION read_files(", "REASON because") {
		t.Errorf("the action line did not immediately follow the reason line; log=%q", log)
	}
}

func TestLogOmitsReasonLineForEscapeOnlyReason(t *testing.T) {
	// An escape-only reason renders no visible text, so it must not emit a
	// dangling reason row. Round 046: the RENDERER decides (the loop honors it);
	// the real sanitize-blanking lives in internal/ui — here we inject a renderer
	// that reports "does not render" for the escape-only value and assert the loop
	// emits no reason line while the action line (preceded by a blank) still does.
	var buf bytes.Buffer
	a := newLogLoop(&buf, llm.ToolCall{ID: "c1", Name: "read_files", Arguments: `{"filepaths":["a.txt"],"reason":"\u001b[31m"}`})
	a.Lines = fakeRenderer{renders: func(string) bool { return false }}
	if _, err := a.Run(context.Background(), "q", nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
	log := buf.String()
	if strings.Contains(log, "REASON ") {
		t.Errorf("an escape-only reason rendered a reason line; log=%q", log)
	}
	if !lineBeforeContains(log, "ACTION read_files(", "") {
		t.Errorf("the reason-less call's action line was not preceded by a blank line; log=%q", log)
	}
}

package agent

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/gosharplite/tellme/internal/domain/llm"
	"github.com/gosharplite/tellme/internal/domain/tools"
)

// Round 039 (operator spacing request; ADR 0008): the loop's begin block for EACH
// tool call is preceded by exactly one blank line — per call, so a k-call round
// emits k blanks. The blank is tied to the call BLOCK, not the reason line, so a
// reason-less call's `[Tool Action]` line is also preceded by a blank.

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
// many are preceded by a blank line. A block's start is its `[Tool Reason]` line
// when the line before the `[Tool Action]` line is a reason line, else the
// `[Tool Action]` line itself.
func blockStartsPrecededByBlank(log string) (total, blank int) {
	lines := logLines(log)
	for i, l := range lines {
		if !strings.Contains(l, "[Tool Action] ") {
			continue
		}
		total++
		begin := i
		if i > 0 && strings.Contains(lines[i-1], "[Tool Reason] ") {
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
	gw := &fakeGateway{responses: []llm.Response{
		{ToolCalls: []llm.ToolCall{
			{ID: "c1", Name: "read_files", Arguments: `{"filepaths":["a.txt"],"reason":"first"}`},
			{ID: "c2", Name: "read_files", Arguments: `{"filepaths":["b.txt"],"reason":"second"}`},
		}},
		{Text: "done"},
	}}
	a := &AgentLoop{
		Gateway:  gw,
		Registry: tools.NewRegistry(fakeTool{name: "read_files", result: "hi"}),
		Stderr:   &buf,
		Now:      fixedLogClock,
	}
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
	gw := &fakeGateway{responses: []llm.Response{
		{ToolCalls: []llm.ToolCall{{ID: "c1", Name: "read_files", Arguments: `{"filepaths":["a.txt"],"reason":"because"}`}}},
		{Text: "done"},
	}}
	a := &AgentLoop{
		Gateway:  gw,
		Registry: tools.NewRegistry(fakeTool{name: "read_files", result: "hi"}),
		Stderr:   &buf,
		Now:      fixedLogClock,
	}
	if _, err := a.Run(context.Background(), "q", nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
	log := buf.String()
	// The blank precedes the REASON line (the block's first line), which is then
	// immediately followed by the action line (no blank between reason and action).
	if !lineBeforeContains(log, "[Tool Reason] because", "") {
		t.Errorf("the reason line was not preceded by a blank line; log=%q", log)
	}
	if !lineBeforeContains(log, "[Tool Action] read_files(", "[12:34:56] [Tool Reason] because") {
		t.Errorf("the action line did not immediately follow the reason line; log=%q", log)
	}
}

func TestLogOmitsReasonLineForEscapeOnlyReason(t *testing.T) {
	// An escape-only reason renders no visible text, so it must not emit a
	// dangling `[Tool Reason]` row (round 039 FR-005) — the action line still
	// starts the block, preceded by a blank.
	var buf bytes.Buffer
	gw := &fakeGateway{responses: []llm.Response{
		{ToolCalls: []llm.ToolCall{{ID: "c1", Name: "read_files", Arguments: `{"filepaths":["a.txt"],"reason":"\u001b[31m"}`}}},
		{Text: "done"},
	}}
	a := &AgentLoop{
		Gateway:  gw,
		Registry: tools.NewRegistry(fakeTool{name: "read_files", result: "hi"}),
		Stderr:   &buf,
		Now:      fixedLogClock,
	}
	if _, err := a.Run(context.Background(), "q", nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
	log := buf.String()
	if strings.Contains(log, "[Tool Reason]") {
		t.Errorf("an escape-only reason rendered a [Tool Reason] line; log=%q", log)
	}
	if !lineBeforeContains(log, "[Tool Action] read_files(", "") {
		t.Errorf("the reason-less call's action line was not preceded by a blank line; log=%q", log)
	}
}

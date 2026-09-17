package agent

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/gosharplite/tellme/internal/domain/llm"
	"github.com/gosharplite/tellme/internal/domain/tools"
)

// Round 036 (issue #74, operator Q2): a blank reason — empty OR whitespace-only
// after folding+trimming — must emit NO `[Tool Reason]` line (and no spurious
// blank line). The loop action line is one of the two suppression sites; the
// per-call tail is pinned in internal/cli. Today the raw `reason != ""` guard
// lets a whitespace-only / newline-only reason through, rendering a dangling row.

func TestLogOmitsReasonLineForWhitespaceOnlyReason(t *testing.T) {
	for _, reason := range []string{"   ", "\n", "\t"} {
		var buf bytes.Buffer
		gw := &fakeGateway{responses: []llm.Response{
			{ToolCalls: []llm.ToolCall{{ID: "call_1", Name: "read_files", Arguments: `{"filepaths":["a.txt"],"reason":"` + reason + `"}`}}},
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
			t.Errorf("a whitespace-only reason %q rendered a [Tool Reason] line; log=%q", reason, log)
		}
		if strings.Contains(log, "\n\n") {
			t.Errorf("a whitespace-only reason %q left a blank line; log=%q", reason, log)
		}
		if !strings.Contains(log, "[Tool Action] read_files(") {
			t.Errorf("the action line was not rendered for reason %q; log=%q", reason, log)
		}
	}
}

func TestLogStillRendersANonBlankReasonLine(t *testing.T) {
	// Control: a real reason still renders exactly one `[Tool Reason]` line, so the
	// suppression is not a blanket "never render".
	var buf bytes.Buffer
	gw := &fakeGateway{responses: []llm.Response{
		{ToolCalls: []llm.ToolCall{{ID: "call_1", Name: "read_files", Arguments: `{"filepaths":["a.txt"],"reason":"because"}`}}},
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
	if log := buf.String(); !strings.Contains(log, "[12:34:56] [Tool Reason] because") {
		t.Errorf("a non-blank reason was dropped; log=%q", log)
	}
}

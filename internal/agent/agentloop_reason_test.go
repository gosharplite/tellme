package agent

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gosharplite/tellme/internal/domain/llm"
	"github.com/gosharplite/tellme/internal/domain/tools"
)

// Round 034: the tool-loop log is the DECOMPOSED shape — `[Tool Engine] Step i/M`,
// `[Tool Reason]`, `[Tool Action] <tool>(<sorted args>)`, `[Tool Result] <tool>:
// <snippet>` (ADR 0005). The injected clock seam keeps the timestamps
// deterministic; the Observer hooks still wrap each write. These pins replace the
// retired round-022 single-line assertions (T003/T004 at the unit layer).

var fixedLogClock = func() time.Time {
	return time.Date(2026, 9, 15, 12, 34, 56, 0, time.UTC)
}

func TestLogRendersDecomposedLinesWithReason(t *testing.T) {
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
	log := buf.String()
	for _, want := range []string{
		"[12:34:56] [Tool Engine] Step 1/1",
		"[12:34:56] [Tool Reason] because",
		`[12:34:56] [Tool Action] read_files(filepaths: ["a.txt"])`,
		"[12:34:56] [Tool Result] read_files: hi",
	} {
		if !strings.Contains(log, want) {
			t.Errorf("tool-loop log missing %q; log=%q", want, log)
		}
	}
	// The reason is excluded from the action's argument list.
	if strings.Contains(log, "reason:") {
		t.Errorf("the action line carried the reason; log=%q", log)
	}
}

func TestLogOmitsReasonLineWhenNoReason(t *testing.T) {
	var buf bytes.Buffer
	gw := &fakeGateway{responses: []llm.Response{
		{ToolCalls: []llm.ToolCall{{ID: "call_1", Name: "read_files", Arguments: `{"filepaths":["a.txt"]}`}}},
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
		t.Errorf("a reasonless call rendered a [Tool Reason] line (FR-006 defensive tolerance); log=%q", log)
	}
	if !strings.Contains(log, "[Tool Action] read_files(") {
		t.Errorf("the action line was not rendered for a reasonless call; log=%q", log)
	}
}

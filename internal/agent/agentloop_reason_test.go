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

// Round 022: the tool-loop log line is reshaped to a single timestamped line
// `[HH:MM:SS] [Tool] <name> - <reason>` (no raw arguments/result); a call with no
// top-level `reason` renders `[HH:MM:SS] [Tool] <name>`. The injected clock seam
// keeps the timestamp deterministic; the Observer hooks still wrap the write.

var fixedLogClock = func() time.Time {
	return time.Date(2026, 9, 15, 12, 34, 56, 0, time.UTC)
}

func TestLogStepRendersToolLineWithReason(t *testing.T) {
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
	if !strings.Contains(log, "[12:34:56] [Tool] read_files - because") {
		t.Errorf("tool-loop log did not render the round-022 line; log=%q", log)
	}
	if strings.Contains(log, "arguments=") || strings.Contains(log, "result=") {
		t.Errorf("tool-loop log echoed the raw arguments/result; log=%q", log)
	}
}

func TestLogStepOmitsReasonWhenAbsent(t *testing.T) {
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
	if !strings.Contains(log, "[12:34:56] [Tool] read_files") {
		t.Errorf("tool-loop log did not name the tool; log=%q", log)
	}
	if strings.Contains(log, " - ") {
		t.Errorf("tool-loop log carried a reason tail when none was given; log=%q", log)
	}
}

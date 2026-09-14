package agent

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/gosharplite/tellme/internal/domain/llm"
	"github.com/gosharplite/tellme/internal/domain/tools"
)

// Round 021 T033: the tool-loop log line echoes a tool call's top-level `reason`
// as `reason=<value>`, and omits the segment when no reason is present (RED until
// Phase 4A). These pin the loop's reason echo — a presentation concern kept out
// of the tool implementations.

func TestLogStepEchoesReason(t *testing.T) {
	var buf bytes.Buffer
	gw := &fakeGateway{responses: []llm.Response{
		{ToolCalls: []llm.ToolCall{{ID: "call_1", Name: "read_files", Arguments: `{"filepaths":["a.txt"],"reason":"because"}`}}},
		{Text: "done"},
	}}
	a := &AgentLoop{Gateway: gw, Registry: tools.NewRegistry(fakeTool{name: "read_files", result: "hi"}), Stderr: &buf}
	if _, err := a.Run(context.Background(), "q", nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(buf.String(), "reason=because") {
		t.Errorf("tool-loop log did not echo reason=because; log=%q", buf.String())
	}
}

func TestLogStepOmitsReasonWhenAbsent(t *testing.T) {
	var buf bytes.Buffer
	gw := &fakeGateway{responses: []llm.Response{
		{ToolCalls: []llm.ToolCall{{ID: "call_1", Name: "read_files", Arguments: `{"filepaths":["a.txt"]}`}}},
		{Text: "done"},
	}}
	a := &AgentLoop{Gateway: gw, Registry: tools.NewRegistry(fakeTool{name: "read_files", result: "hi"}), Stderr: &buf}
	if _, err := a.Run(context.Background(), "q", nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if strings.Contains(buf.String(), "reason=") {
		t.Errorf("tool-loop log carried a reason segment when none was given; log=%q", buf.String())
	}
}

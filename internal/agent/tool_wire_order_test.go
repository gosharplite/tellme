package agent

import (
	"context"
	"testing"

	"github.com/gosharplite/tellme/internal/domain/llm"
	"github.com/gosharplite/tellme/internal/domain/tools"
)

// TDD regression (review PR #25 BLOCKER-1): the tool-loop wire chronology must
// be `user(prompt) → assistant(tool_calls) → tool(result)`. The current prompt
// must NOT be sent after the tool activity.
func TestRunWireChronology(t *testing.T) {
	gw := &fakeGateway{responses: []llm.Response{
		{ToolCalls: []llm.ToolCall{{ID: "call_1", Name: "read_files", Arguments: `{"path":"x"}`}}},
		{Text: "done"},
	}}
	a := &AgentLoop{Gateway: gw, Registry: tools.NewRegistry(fakeTool{name: "read_files", result: "R"})}
	if _, err := a.Run(context.Background(), "read x", nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(gw.calls) < 2 {
		t.Fatalf("provider calls = %d, want >= 2", len(gw.calls))
	}
	// Iteration 0: the round-004/007 shape — the prompt is the request's user
	// message and Messages is the (empty) prior conversation.
	if first := gw.calls[0]; first.Prompt != "read x" || len(first.Messages) != 0 {
		t.Fatalf("first request = %+v, want prompt only", first)
	}
	// Iteration 1: the active turn folded into Messages, chronologically ordered.
	second := gw.calls[1]
	if second.Prompt != "" {
		t.Errorf("second request Prompt = %q, want empty (folded into messages)", second.Prompt)
	}
	if len(second.Messages) != 3 {
		t.Fatalf("second request messages = %d, want 3: %+v", len(second.Messages), second.Messages)
	}
	if second.Messages[0].Role != "user" || second.Messages[0].Content != "read x" {
		t.Errorf("messages[0] = %+v, want user/read x", second.Messages[0])
	}
	if second.Messages[1].Role != "assistant" || len(second.Messages[1].ToolCalls) != 1 {
		t.Errorf("messages[1] = %+v, want assistant with one tool_call", second.Messages[1])
	}
	if second.Messages[2].Role != "tool" || second.Messages[2].ToolCallID != "call_1" {
		t.Errorf("messages[2] = %+v, want tool/call_1", second.Messages[2])
	}
}

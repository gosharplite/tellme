package agent

import (
	"context"
	"testing"

	"github.com/gosharplite/tellme/internal/domain/llm"
	"github.com/gosharplite/tellme/internal/domain/tools"
)

// Round-018 UNIT (T021): the loop accumulates every provider call's usage on
// AgentResult.Calls (in call order), so the CLI can compute the turn cost
// ($#2 = the sum) while AgentResult.Usage stays the just-returned (final) call.
func TestRunAccumulatesCallUsage(t *testing.T) {
	u1 := llm.Usage{Reported: true, PromptTokens: 10, CachedTokens: 6, CompletionTokens: 3, ThinkingTokens: 2}
	u2 := llm.Usage{Reported: true, PromptTokens: 12, CachedTokens: 8, CompletionTokens: 4, ThinkingTokens: 1}
	gw := &fakeGateway{responses: []llm.Response{
		{ToolCalls: []llm.ToolCall{{ID: "call_1", Name: "read_files", Arguments: `{"path":"notes.txt"}`}}, Usage: u1},
		{Text: "done", Usage: u2},
	}}
	a := &AgentLoop{Gateway: gw, Registry: tools.NewRegistry(fakeTool{name: "read_files", result: "x"})}
	res, err := a.Run(context.Background(), "read notes.txt", nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.Calls) != 2 || res.Calls[0] != u1 || res.Calls[1] != u2 {
		t.Fatalf("Calls = %+v, want [u1 u2]", res.Calls)
	}
	if res.Usage != u2 {
		t.Errorf("Usage = %+v, want the just-returned (last) call u2", res.Usage)
	}
}

// A single-call turn records exactly one usage entry.
func TestRunOneCallUsage(t *testing.T) {
	u := llm.Usage{Reported: true, PromptTokens: 5, CachedTokens: 0, CompletionTokens: 7, ThinkingTokens: 0}
	gw := &fakeGateway{responses: []llm.Response{{Text: "hi", Usage: u}}}
	a := &AgentLoop{Gateway: gw, Registry: tools.NewRegistry()}
	res, err := a.Run(context.Background(), "q", nil)
	if err != nil || len(res.Calls) != 1 || res.Calls[0] != u {
		t.Fatalf("Run = (%+v, %v), want one usage entry %+v", res.Calls, err, u)
	}
}

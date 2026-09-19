package agent

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/gosharplite/tellme/internal/domain/llm"
	"github.com/gosharplite/tellme/internal/domain/tools"
)

// T014 [UNIT] — round 056 (ADR 0025 D3; folds #121): the universal *no reason, no
// go* gate. A call whose reason does not render is REFUSED — the tool does not
// execute — and the loop folds back a recoverable result asking the model to
// retry with a reason.
//
// The decision REUSES the round-046 single owner: the loop asks its injected
// `Lines.ReasonLine` port (no second predicate). This pin drives the decision
// through a fake renderer that suppresses, so it proves the loop HONORS the
// owner's `renders` bool rather than inspecting the raw reason itself.
func TestRunRefusesCallWithoutRenderableReason(t *testing.T) {
	gw := &fakeGateway{responses: []llm.Response{
		{ToolCalls: []llm.ToolCall{{ID: "call_1", Name: "read_files", Arguments: `{"filepaths":["notes.txt"]}`}}},
		{Text: "done"},
	}}
	tool := &countingTool{name: "read_files", result: "content"}
	a := &AgentLoop{
		Gateway:  gw,
		Registry: tools.NewRegistry(tool),
		// The owner reports the reason as non-rendering (a reason-less call, or an
		// escape-only reason the real sanitizer would blank).
		Lines: fakeRenderer{renders: func(string) bool { return false }},
	}
	res, err := a.Run(context.Background(), "read notes.txt", nil)
	if err != nil {
		t.Fatalf("a refusal must not fail the run: %v", err)
	}
	if tool.executions != 0 {
		t.Fatalf("a reason-less call must NOT execute the tool; executions=%d", tool.executions)
	}
	if len(res.Steps) != 0 {
		t.Fatalf("a refused call is not an executed step; steps=%+v", res.Steps)
	}
	// The recoverable result must have been fed back (a `tool`-role message).
	if !refusalFedBack(gw) {
		t.Fatal("the refusal result was not fed back to the model")
	}
	found := false
	for _, m := range gw.calls[1].Messages {
		if m.Role == "tool" && strings.Contains(m.Content, "a reason is required") {
			found = true
		}
	}
	if !found {
		t.Fatalf("the recoverable `a reason is required` result must be a tool message; messages=%+v", gw.calls[1].Messages)
	}
}

// T014 [UNIT] — a call WITH a renderable reason executes normally (the gate is a
// no-op on the happy path).
func TestRunExecutesCallWithRenderableReason(t *testing.T) {
	gw := &fakeGateway{responses: []llm.Response{
		{ToolCalls: []llm.ToolCall{{ID: "call_1", Name: "read_files", Arguments: `{"filepaths":["notes.txt"],"reason":"inspect notes"}`}}},
		{Text: "done"},
	}}
	tool := &countingTool{name: "read_files", result: "content"}
	a := &AgentLoop{Gateway: gw, Registry: tools.NewRegistry(tool), Lines: fakeRenderer{}}
	if _, err := a.Run(context.Background(), "read notes.txt", nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if tool.executions != 1 {
		t.Fatalf("a call stating a reason must execute once; executions=%d", tool.executions)
	}
}

// T014 [UNIT] — the ADR-0025 boundary: a NIL renderer means NO gate (there is no
// predicate to ask). This preserves the round-031 assembler gate / the offline
// `--tool-usage` path, where no prompt turn runs.
func TestRunNilRendererDoesNotGate(t *testing.T) {
	gw := &fakeGateway{responses: []llm.Response{
		{ToolCalls: []llm.ToolCall{{ID: "call_1", Name: "read_files", Arguments: `{"filepaths":["notes.txt"]}`}}},
		{Text: "done"},
	}}
	tool := &countingTool{name: "read_files", result: "content"}
	a := &AgentLoop{Gateway: gw, Registry: tools.NewRegistry(tool)} // Lines nil
	if _, err := a.Run(context.Background(), "read notes.txt", nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if tool.executions != 1 {
		t.Fatalf("with no renderer the gate must not apply; executions=%d", tool.executions)
	}
}

// countingTool is a fake tool that counts executions.
type countingTool struct {
	name       string
	result     string
	executions int
}

func (c *countingTool) Name() string        { return c.name }
func (c *countingTool) Description() string { return "fake tool" }
func (c *countingTool) Parameters() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{}}`)
}
func (c *countingTool) Contract() tools.ToolContract { return tools.ToolContract{} }
func (c *countingTool) Execute(context.Context, string, tools.ByteBudget) (string, error) {
	c.executions++
	return c.result, nil
}

// refusalFedBack reports whether any request after the first carried a tool
// message.
func refusalFedBack(gw *fakeGateway) bool {
	for _, req := range gw.calls[1:] {
		for _, m := range req.Messages {
			if m.Role == "tool" {
				return true
			}
		}
	}
	return false
}

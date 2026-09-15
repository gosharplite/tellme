package agent

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/gosharplite/tellme/internal/domain/history"
	"github.com/gosharplite/tellme/internal/domain/llm"
	"github.com/gosharplite/tellme/internal/domain/tools"
)

// fakeGateway replays a scripted sequence of responses.
type fakeGateway struct {
	responses []llm.Response
	calls     []llm.Request
	i         int
}

func (f *fakeGateway) Complete(_ context.Context, req llm.Request) (llm.Response, error) {
	f.calls = append(f.calls, req)
	if f.i >= len(f.responses) {
		return llm.Response{}, errors.New("no more scripted responses")
	}
	r := f.responses[f.i]
	f.i++
	return r, nil
}

// fakeTool is a canned read-only tool.
type fakeTool struct {
	name   string
	result string
	err    error
}

func (f fakeTool) Name() string        { return f.name }
func (f fakeTool) Description() string { return "fake tool" }
func (f fakeTool) Parameters() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{}}`)
}
func (f fakeTool) Contract() tools.ToolContract { return tools.ToolContract{} }
func (f fakeTool) Execute(context.Context, string, tools.ByteBudget) (string, error) {
	return f.result, f.err
}

// TestRunNoToolCalls: a plain answer makes one request and no tool steps.
func TestRunNoToolCalls(t *testing.T) {
	gw := &fakeGateway{responses: []llm.Response{{Text: "hi"}}}
	a := &AgentLoop{Gateway: gw, Registry: tools.NewRegistry()}
	res, err := a.Run(context.Background(), "q", nil)
	if err != nil || res.Answer != "hi" || len(res.Steps) != 0 {
		t.Fatalf("Run = (%q, %v, %v), want (hi, 0 steps, nil)", res.Answer, res.Steps, err)
	}
	if len(gw.calls) != 1 {
		t.Errorf("provider calls = %d, want 1", len(gw.calls))
	}
}

// TestRunOneToolRound: one tool round then a final answer.
func TestRunOneToolRound(t *testing.T) {
	gw := &fakeGateway{responses: []llm.Response{
		{ToolCalls: []llm.ToolCall{{ID: "call_1", Name: "read_files", Arguments: `{"path":"notes.txt"}`}}},
		{Text: "the launch code is ORANGE"},
	}}
	a := &AgentLoop{Gateway: gw, Registry: tools.NewRegistry(fakeTool{name: "read_files", result: "the launch code is ORANGE"})}
	res, err := a.Run(context.Background(), "read notes.txt", nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Answer != "the launch code is ORANGE" || len(res.Steps) != 1 || res.Steps[0].Tool != "read_files" {
		t.Fatalf("Run = (%q, %+v)", res.Answer, res.Steps)
	}
}

// TestRunToolErrorIsFedBack: a tool error is non-terminal (fed back, run finishes).
func TestRunToolErrorIsFedBack(t *testing.T) {
	gw := &fakeGateway{responses: []llm.Response{
		{ToolCalls: []llm.ToolCall{{ID: "call_1", Name: "read_files", Arguments: `{"path":"missing"}`}}},
		{Text: "no such file"},
	}}
	a := &AgentLoop{Gateway: gw, Registry: tools.NewRegistry(fakeTool{name: "read_files", err: errors.New("open missing: no such file")})}
	res, err := a.Run(context.Background(), "read missing", nil)
	if err != nil || res.Answer != "no such file" || len(res.Steps) != 1 {
		t.Fatalf("Run = (%q, %+v, %v)", res.Answer, res.Steps, err)
	}
	if len(gw.calls) < 2 {
		t.Fatalf("expected the error to be fed back into a second call")
	}
}

// TestRunBoundReached: a model that always requests tools stops at MaxLoops.
func TestRunBoundReached(t *testing.T) {
	tc := llm.ToolCall{ID: "c", Name: "read_files", Arguments: "{}"}
	gw := &fakeGateway{responses: []llm.Response{{ToolCalls: []llm.ToolCall{tc}}, {ToolCalls: []llm.ToolCall{tc}}, {ToolCalls: []llm.ToolCall{tc}}}}
	a := &AgentLoop{Gateway: gw, Registry: tools.NewRegistry(fakeTool{name: "read_files", result: "x"}), MaxLoops: 2}
	res, err := a.Run(context.Background(), "loop", nil)
	var inc *ErrIncomplete
	if !errors.As(err, &inc) {
		t.Fatalf("err = %v, want *ErrIncomplete", err)
	}
	if len(res.Steps) != 2 {
		t.Errorf("steps = %d, want 2 (the bound)", len(res.Steps))
	}
}

// TestRunUnknownToolIsTerminal: an undeclared tool fails fast.
func TestRunUnknownToolIsTerminal(t *testing.T) {
	gw := &fakeGateway{responses: []llm.Response{{ToolCalls: []llm.ToolCall{{ID: "c", Name: "time_travel"}}}}}
	a := &AgentLoop{Gateway: gw, Registry: tools.NewRegistry()}
	_, err := a.Run(context.Background(), "use the time-travel tool", nil)
	var inc *ErrIncomplete
	if !errors.As(err, &inc) {
		t.Fatalf("err = %v, want *ErrIncomplete", err)
	}
}

// TestBuildMessagesReplaysToolSteps pins the deterministic replay ids (TD-2).
func TestBuildMessagesReplaysToolSteps(t *testing.T) {
	prior := []history.Entry{{
		Prompt: "read notes.txt",
		Answer: "ORANGE",
		Steps:  []history.Step{{Tool: "read_files", Arguments: `{"path":"notes.txt"}`, Result: "ORANGE"}},
	}}
	msgs := BuildMessages(prior)
	if len(msgs) != 4 {
		t.Fatalf("messages = %d, want 4", len(msgs))
	}
	if msgs[0].Role != "user" || msgs[3].Role != "assistant" || msgs[3].Content != "ORANGE" {
		t.Fatalf("messages = %+v", msgs)
	}
	if msgs[1].Role != "assistant" || len(msgs[1].ToolCalls) != 1 || msgs[1].ToolCalls[0].ID != "call_step_1" {
		t.Fatalf("replay tool-call = %+v, want id call_step_1", msgs[1])
	}
	if msgs[2].Role != "tool" || msgs[2].ToolCallID != "call_step_1" {
		t.Fatalf("replay tool result = %+v", msgs[2])
	}
}

// T008 (round 014) — the persisted provider token replays into the synthesised
// tool call (the deterministic id is unchanged).
func TestBuildMessagesReplaysSignature(t *testing.T) {
	prior := []history.Entry{{
		Prompt: "read notes.txt",
		Answer: "ORANGE",
		Steps:  []history.Step{{Tool: "read_files", Arguments: `{"path":"notes.txt"}`, Result: "ORANGE", Signature: "sig-abc"}},
	}}
	msgs := BuildMessages(prior)
	if len(msgs) != 4 || len(msgs[1].ToolCalls) != 1 {
		t.Fatalf("messages = %+v", msgs)
	}
	if msgs[1].ToolCalls[0].Signature != "sig-abc" {
		t.Fatalf("replayed tool-call signature = %q, want sig-abc", msgs[1].ToolCalls[0].Signature)
	}
	if msgs[1].ToolCalls[0].ID != "call_step_1" {
		t.Fatalf("replayed tool-call id = %q, want call_step_1", msgs[1].ToolCalls[0].ID)
	}
}

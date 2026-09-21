package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	agentport "github.com/gosharplite/tellme/internal/domain/agent"
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
	var inc *agentport.ErrIncomplete
	if !errors.As(err, &inc) {
		t.Fatalf("err = %v, want *agentport.ErrIncomplete", err)
	}
	if len(res.Steps) != 2 {
		t.Errorf("steps = %d, want 2 (the bound)", len(res.Steps))
	}
}

// TestRunUnknownToolIsFedBackAndContinues: an undeclared tool name is a
// recoverable slip — the loop folds back a `tool`-role result and continues to a
// final answer (round 076). Folds F-1 (the pairing) + TD-076-1 (the list).
func TestRunUnknownToolIsFedBackAndContinues(t *testing.T) {
	gw := &fakeGateway{responses: []llm.Response{
		{ToolCalls: []llm.ToolCall{{ID: "c1", Name: "time_travel"}}},
		{Text: "done"},
	}}
	a := &AgentLoop{Gateway: gw, Registry: tools.NewRegistry(fakeTool{name: "read_files", result: "x"})}
	res, err := a.Run(context.Background(), "use the time-travel tool", nil)
	if err != nil {
		t.Fatalf("an unknown tool name must not fail the run: %v", err)
	}
	if res.Answer != "done" {
		t.Fatalf("answer = %q, want the loop to continue to a final answer", res.Answer)
	}
	if len(res.Steps) != 0 {
		t.Fatalf("an unknown call is not an executed step; steps=%+v", res.Steps)
	}
	if len(gw.calls) < 2 {
		t.Fatalf("expected the recoverable result to be fed back into a second call; calls=%d", len(gw.calls))
	}
	// F-1: the folded-back `tool` message MUST carry the unknown call's id (the
	// Gemini/Vertex function-call/response pairing — round-065 / #132).
	var folded *llm.Message
	for i := range gw.calls[1].Messages {
		if m := &gw.calls[1].Messages[i]; m.Role == "tool" {
			folded = m
		}
	}
	if folded == nil {
		t.Fatalf("no `tool`-role fold-back was fed back; messages=%+v", gw.calls[1].Messages)
	}
	if folded.ToolCallID != "c1" {
		t.Fatalf("the fold-back MUST pair to the call id; ToolCallID = %q, want %q", folded.ToolCallID, "c1")
	}
	// TD-076-1: the message names the unknown tool AND lists the available tools.
	if !strings.Contains(folded.Content, `no tool named "time_travel"`) {
		t.Fatalf("the fold-back must name the unknown tool; content=%q", folded.Content)
	}
	if !strings.Contains(folded.Content, "available tools:") || !strings.Contains(folded.Content, "read_files") {
		t.Fatalf("the fold-back must list the available wire names; content=%q", folded.Content)
	}
}

// TestRunUnknownToolRecordsNoUsage: an unknown call executes nothing and records
// NO tool-usage entry (round 076 fold F-1 — the I-2 no-usage half).
func TestRunUnknownToolRecordsNoUsage(t *testing.T) {
	sink := &recordingSink{}
	gw := &fakeGateway{responses: []llm.Response{
		{ToolCalls: []llm.ToolCall{{ID: "c1", Name: "time_travel"}}},
		{Text: "done"},
	}}
	a := &AgentLoop{Gateway: gw, Registry: tools.NewRegistry(fakeTool{name: "read_files", result: "x"}), ToolUsage: sink}
	if _, err := a.Run(context.Background(), "use the time-travel tool", nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(sink.tools) != 0 {
		t.Fatalf("an unknown call must record no usage; recorded %v", sink.tools)
	}
}

// TestRunMixedRoundUnknownAndValid: in ONE round carrying an unknown AND a valid
// call, every call gets a paired result and the valid call still executes
// (round 076 fold F-1 — the pairing invariant across a mixed round).
func TestRunMixedRoundUnknownAndValid(t *testing.T) {
	gw := &fakeGateway{responses: []llm.Response{
		{ToolCalls: []llm.ToolCall{
			{ID: "c1", Name: "time_travel"},
			{ID: "c2", Name: "read_files", Arguments: `{"filepaths":["notes.txt"],"reason":"read it"}`},
		}},
		{Text: "done"},
	}}
	tool := &countingTool{name: "read_files", result: "content"}
	a := &AgentLoop{Gateway: gw, Registry: tools.NewRegistry(tool), Lines: fakeRenderer{}}
	res, err := a.Run(context.Background(), "read the note", nil)
	if err != nil {
		t.Fatalf("a mixed round must not fail the run: %v", err)
	}
	if tool.executions != 1 {
		t.Fatalf("the valid call must still execute once; executions=%d", tool.executions)
	}
	if len(res.Steps) != 1 || res.Steps[0].Tool != "read_files" {
		t.Fatalf("only the valid call is a step; steps=%+v", res.Steps)
	}
	ids := map[string]bool{}
	for _, m := range gw.calls[1].Messages {
		if m.Role == "tool" {
			ids[m.ToolCallID] = true
		}
	}
	if !ids["c1"] || !ids["c2"] {
		t.Fatalf("every call must get a paired `tool` result; tool ids=%v", ids)
	}
}

// TestRunUnknownToolWithBlankReasonClassifiesAsUnknown: an unknown call that also
// lacks a renderable reason is classified as UNKNOWN (the lookup precedes the
// reason gate), single-owned — round 076 fold F-3.
func TestRunUnknownToolWithBlankReasonClassifiesAsUnknown(t *testing.T) {
	gw := &fakeGateway{responses: []llm.Response{
		{ToolCalls: []llm.ToolCall{{ID: "c1", Name: "time_travel", Arguments: `{}`}}},
		{Text: "done"},
	}}
	a := &AgentLoop{
		Gateway:  gw,
		Registry: tools.NewRegistry(fakeTool{name: "read_files", result: "x"}),
		Lines:    fakeRenderer{renders: func(string) bool { return false }},
	}
	if _, err := a.Run(context.Background(), "use the time-travel tool", nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, m := range gw.calls[1].Messages {
		if m.Role != "tool" {
			continue
		}
		if strings.Contains(m.Content, "a reason is required") {
			t.Fatalf("an unknown name must classify as unknown, not reason-less; content=%q", m.Content)
		}
		if strings.Contains(m.Content, `no tool named "time_travel"`) {
			return
		}
	}
	t.Fatalf("no unknown-name fold-back was fed back; messages=%+v", gw.calls[1].Messages)
}

// TestRunUnknownFoldBackEmitsActionLine: the fold-back mirrors the reason-less
// refusal's chrome — it emits the call's `[Tool Action]` block before the result
// (round 076 fold F-5).
func TestRunUnknownFoldBackEmitsActionLine(t *testing.T) {
	gw := &fakeGateway{responses: []llm.Response{
		{ToolCalls: []llm.ToolCall{{ID: "c1", Name: "time_travel"}}},
		{Text: "done"},
	}}
	var buf bytes.Buffer
	a := &AgentLoop{Gateway: gw, Registry: tools.NewRegistry(fakeTool{name: "read_files", result: "x"}), Stderr: &buf, Lines: fakeRenderer{}}
	if _, err := a.Run(context.Background(), "use the time-travel tool", nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(buf.String(), "ACTION time_travel") {
		t.Fatalf("the fold-back must emit the call's action line; stderr=%q", buf.String())
	}
	if !strings.Contains(buf.String(), "RESULT time_travel:") {
		t.Fatalf("the fold-back must emit the result line; stderr=%q", buf.String())
	}
}

// TestRunUnknownFoldBackCounterResetsPerRun: the per-turn cap counter is Run-local
// — a second Run on the same loop starts with a fresh budget (round 076 fold
// TD-076-2).
func TestRunUnknownFoldBackCounterResetsPerRun(t *testing.T) {
	// One unknown call then an answer — well under the cap — repeated twice on the
	// SAME loop: if the counter leaked across Runs the second Run would still be
	// fine (1+1 < 3), so drive the loop to exactly the cap once, then a fresh Run.
	unknownThenAnswer := []llm.Response{
		{ToolCalls: []llm.ToolCall{{ID: "c1", Name: "time_travel"}}},
		{ToolCalls: []llm.ToolCall{{ID: "c2", Name: "time_travel"}}},
		{ToolCalls: []llm.ToolCall{{ID: "c3", Name: "time_travel"}}},
		{Text: "done"},
	}
	gw := &fakeGateway{responses: unknownThenAnswer}
	a := &AgentLoop{Gateway: gw, Registry: tools.NewRegistry(fakeTool{name: "read_files", result: "x"}), MaxLoops: 100}
	if _, err := a.Run(context.Background(), "use it", nil); err != nil {
		t.Fatalf("first Run (exactly the cap) must succeed: %v", err)
	}
	// A second Run reuses the loop; a leaked counter would immediately abort at the
	// first unknown call. Reset the gateway cursor and answer script.
	gw.i = 0
	gw.calls = nil
	gw2 := &fakeGateway{responses: []llm.Response{
		{ToolCalls: []llm.ToolCall{{ID: "d1", Name: "time_travel"}}},
		{Text: "done"},
	}}
	a.Gateway = gw2
	if _, err := a.Run(context.Background(), "use it", nil); err != nil {
		t.Fatalf("a second Run must reset the per-turn cap counter: %v", err)
	}
}

// TestMaxUnknownToolFoldsValuePinned pins the per-turn cap VALUE (round 076 fold
// F-2): the recoverable fold-back is bounded at 3. Changing it is a deliberate,
// reviewed act — this pin forces the reviewer to see the value change.
func TestMaxUnknownToolFoldsValuePinned(t *testing.T) {
	if maxUnknownToolFolds != 3 {
		t.Fatalf("maxUnknownToolFolds = %d, want the pinned value 3 (round 076 / ADR 0048)", maxUnknownToolFolds)
	}
}

// TestRunUnknownToolIsBoundedPerTurn: a provider that keeps asking for an unknown
// tool is stopped at the per-turn cap (round 076) — a bounded fold-back, not an
// unbounded spin. The fake is scripted with MORE replies than the cap so raising
// the constant fails on the cap assertion, not on fake exhaustion (fold TD-076-4).
func TestRunUnknownToolIsBoundedPerTurn(t *testing.T) {
	tc := llm.ToolCall{ID: "c", Name: "time_travel"}
	replies := make([]llm.Response, 0, 20)
	for range 20 {
		replies = append(replies, llm.Response{ToolCalls: []llm.ToolCall{tc}})
	}
	gw := &fakeGateway{responses: replies}
	a := &AgentLoop{Gateway: gw, Registry: tools.NewRegistry(fakeTool{name: "read_files", result: "x"}), MaxLoops: 100}
	_, err := a.Run(context.Background(), "use the time-travel tool", nil)
	var inc *agentport.ErrIncomplete
	if !errors.As(err, &inc) {
		t.Fatalf("err = %v, want *agentport.ErrIncomplete after the per-turn cap", err)
	}
	// The cap allows maxUnknownToolFolds recoverable rounds, then the (N+1)th
	// request is terminal: the provider was called N+1 times.
	if len(gw.calls) != maxUnknownToolFolds+1 {
		t.Fatalf("provider calls = %d, want %d (the per-turn cap)", len(gw.calls), maxUnknownToolFolds+1)
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

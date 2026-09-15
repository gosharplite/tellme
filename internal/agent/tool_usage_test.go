package agent

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/gosharplite/tellme/internal/domain/history"
	"github.com/gosharplite/tellme/internal/domain/llm"
	"github.com/gosharplite/tellme/internal/domain/tools"
)

// recordingSink captures the recorded (tool, outcome) pairs in call order.
type recordingSink struct {
	tools    []string
	outcomes []history.ToolOutcome
}

func (s *recordingSink) Record(tool string, outcome history.ToolOutcome) error {
	s.tools = append(s.tools, tool)
	s.outcomes = append(s.outcomes, outcome)
	return nil
}

// erroringSink always fails; the loop must swallow it (best-effort).
type erroringSink struct{ calls int }

func (s *erroringSink) Record(string, history.ToolOutcome) error {
	s.calls++
	return errors.New("sink unavailable")
}

// blockingTool blocks until its per-call deadline expires, then returns the
// round-024 FR-018 nil-error timeout result.
type blockingTool struct {
	name   string
	result string
}

func (b blockingTool) Name() string        { return b.name }
func (b blockingTool) Description() string { return "blocking tool" }
func (b blockingTool) Parameters() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{}}`)
}
func (b blockingTool) Contract() tools.ToolContract {
	return tools.ToolContract{DefaultTimeout: 20 * time.Millisecond}
}
func (b blockingTool) Execute(ctx context.Context, _ string, _ tools.ByteBudget) (string, error) {
	<-ctx.Done()
	return b.result, nil
}

// TestClassifyToolOutcome is the pure classification table: the outcome follows
// the loop's structural signals (err + the deadline), with error taking
// precedence over a simultaneously-expired deadline.
func TestClassifyToolOutcome(t *testing.T) {
	cases := []struct {
		name     string
		terr     error
		timedOut bool
		want     history.ToolOutcome
	}{
		{"ok", nil, false, history.ToolOutcomeOK},
		{"error", errors.New("boom"), false, history.ToolOutcomeError},
		{"timeout", nil, true, history.ToolOutcomeTimeout},
		{"error wins over a fired deadline", errors.New("boom"), true, history.ToolOutcomeError},
	}
	for _, c := range cases {
		if got := classifyToolOutcome(c.terr, c.timedOut); got != c.want {
			t.Errorf("%s: classifyToolOutcome = %q, want %q", c.name, got, c.want)
		}
	}
}

// TestRecordToolUsageOK: a tool that returns a result is recorded `ok`.
func TestRecordToolUsageOK(t *testing.T) {
	sink := &recordingSink{}
	gw := &fakeGateway{responses: []llm.Response{
		{ToolCalls: []llm.ToolCall{{ID: "call_1", Name: "read_files", Arguments: "{}"}}},
		{Text: "done"},
	}}
	a := &AgentLoop{Gateway: gw, Registry: tools.NewRegistry(fakeTool{name: "read_files", result: "contents"}), ToolUsage: sink}
	if _, err := a.Run(context.Background(), "q", nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(sink.outcomes) != 1 || sink.outcomes[0] != history.ToolOutcomeOK || sink.tools[0] != "read_files" {
		t.Fatalf("recorded %v / %v, want [read_files] / [ok]", sink.tools, sink.outcomes)
	}
}

// TestRecordToolUsageError: a tool that returns an error is recorded `error`.
func TestRecordToolUsageError(t *testing.T) {
	sink := &recordingSink{}
	gw := &fakeGateway{responses: []llm.Response{
		{ToolCalls: []llm.ToolCall{{ID: "call_1", Name: "list_files", Arguments: "{}"}}},
		{Text: "done"},
	}}
	a := &AgentLoop{Gateway: gw, Registry: tools.NewRegistry(fakeTool{name: "list_files", err: errors.New("no such directory")}), ToolUsage: sink}
	if _, err := a.Run(context.Background(), "q", nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(sink.outcomes) != 1 || sink.outcomes[0] != history.ToolOutcomeError {
		t.Fatalf("recorded %v, want [error]", sink.outcomes)
	}
}

// TestRecordToolUsageTimeoutLeg: a tool stopped at its per-call deadline (a
// nil-error result) is recorded `timeout` — the FR-018 leg.
func TestRecordToolUsageTimeoutLeg(t *testing.T) {
	sink := &recordingSink{}
	gw := &fakeGateway{responses: []llm.Response{
		{ToolCalls: []llm.ToolCall{{ID: "call_1", Name: "execute_command", Arguments: "{}"}}},
		{Text: "done"},
	}}
	a := &AgentLoop{Gateway: gw, Registry: tools.NewRegistry(blockingTool{name: "execute_command", result: "\n... (stopped at the time limit)\n"}), ToolUsage: sink}
	if _, err := a.Run(context.Background(), "q", nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(sink.outcomes) != 1 || sink.outcomes[0] != history.ToolOutcomeTimeout {
		t.Fatalf("recorded %v, want [timeout]", sink.outcomes)
	}
}

// TestRecordToolUsageTieBreakFollowsDeadline pins the trim-vs-deadline tie-break:
// a tool whose RESULT text is a bounded success (a truncation marker) but whose
// per-call deadline expired is recorded `timeout` — the accounting follows the
// loop's deadline signal, not the result shape (round-026 review F3).
func TestRecordToolUsageTieBreakFollowsDeadline(t *testing.T) {
	sink := &recordingSink{}
	gw := &fakeGateway{responses: []llm.Response{
		{ToolCalls: []llm.ToolCall{{ID: "call_1", Name: "execute_command", Arguments: "{}"}}},
		{Text: "done"},
	}}
	// The result text looks like a trim, yet the deadline fired.
	a := &AgentLoop{Gateway: gw, Registry: tools.NewRegistry(blockingTool{name: "execute_command", result: "partial output" + tools.TruncationMarker}), ToolUsage: sink}
	if _, err := a.Run(context.Background(), "q", nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(sink.outcomes) != 1 || sink.outcomes[0] != history.ToolOutcomeTimeout {
		t.Fatalf("tie-break recorded %v, want [timeout] (the loop's deadline signal)", sink.outcomes)
	}
}

// TestRecordToolUsageSinkErrorSwallowed: a failing sink never breaks the turn.
func TestRecordToolUsageSinkErrorSwallowed(t *testing.T) {
	sink := &erroringSink{}
	gw := &fakeGateway{responses: []llm.Response{
		{ToolCalls: []llm.ToolCall{{ID: "call_1", Name: "read_files", Arguments: "{}"}}},
		{Text: "done"},
	}}
	a := &AgentLoop{Gateway: gw, Registry: tools.NewRegistry(fakeTool{name: "read_files", result: "contents"}), ToolUsage: sink}
	res, err := a.Run(context.Background(), "q", nil)
	if err != nil || res.Answer != "done" {
		t.Fatalf("Run = (%q, %v), want (done, nil); a sink error must be swallowed", res.Answer, err)
	}
	if sink.calls != 1 {
		t.Errorf("sink calls = %d, want 1", sink.calls)
	}
}

// TestRecordToolUsageNilSinkNoOp: a nil sink is a no-op (no panic).
func TestRecordToolUsageNilSinkNoOp(t *testing.T) {
	gw := &fakeGateway{responses: []llm.Response{
		{ToolCalls: []llm.ToolCall{{ID: "call_1", Name: "read_files", Arguments: "{}"}}},
		{Text: "done"},
	}}
	a := &AgentLoop{Gateway: gw, Registry: tools.NewRegistry(fakeTool{name: "read_files", result: "contents"})}
	if _, err := a.Run(context.Background(), "q", nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
}

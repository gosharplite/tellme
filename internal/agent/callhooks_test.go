package agent

import (
	"context"
	"testing"

	agentport "github.com/gosharplite/tellme/internal/domain/agent"
	"github.com/gosharplite/tellme/internal/domain/llm"
	"github.com/gosharplite/tellme/internal/domain/tools"
)

// Round-034 T022 unit pins for the per-call estimator seam (ADR 0005 D1/D2): the
// loop fires the call-begin hook with the FUSED base+turn wire messages (no
// loop-owned estimator field), so the CLI can compute
// `llm.EstimatePayload(person, agentport.ToolDefs(reg), wire_k)` — call 1 byte-identical to
// the previous once-per-prompt estimate, calls 2..k growing monotonically.

// recordingObserver records the round-034 call hooks; the waiting-phase hooks are
// no-ops (this round does not exercise the spinner here).
type recordingObserver struct {
	begins []beginRec
	ends   []endRec
}

type beginRec struct {
	idx  int
	msgs []llm.Message
}

type endRec struct {
	idx     int
	usage   llm.Usage
	reasons []string
	final   bool
}

func (o *recordingObserver) OnCallBegin(i int, m []llm.Message) {
	o.begins = append(o.begins, beginRec{i, m})
}

func (o *recordingObserver) OnCallEnd(i int, u llm.Usage, r []string, f bool) {
	o.ends = append(o.ends, endRec{i, u, r, f})
}

func (o *recordingObserver) OnInferenceStart()     {}
func (o *recordingObserver) OnInferenceEnd()       {}
func (o *recordingObserver) OnToolsStart([]string) {}
func (o *recordingObserver) OnToolsEnd()           {}
func (o *recordingObserver) YieldIndicator()       {}
func (o *recordingObserver) RestoreIndicator()     {}

func TestAgentLoop_FiresCallHooksWithFusedMessages(t *testing.T) {
	gw := &fakeGateway{responses: []llm.Response{
		{
			ToolCalls: []llm.ToolCall{{ID: "c1", Name: "read_files", Arguments: `{"filepaths":["notes.txt"],"reason":"checking"}`}},
			Usage:     llm.Usage{PromptTokens: 100, Reported: true},
		},
		{Text: "ORANGE", Usage: llm.Usage{PromptTokens: 250, Reported: true}},
	}}
	reg := tools.NewRegistry(fakeTool{name: "read_files", result: "ORANGE"})
	obs := &recordingObserver{}
	a := &AgentLoop{Gateway: gw, Registry: reg, Observer: obs, Lines: fakeRenderer{}}

	if _, err := a.Run(context.Background(), "read notes.txt", nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(obs.begins) != 2 || len(obs.ends) != 2 {
		t.Fatalf("call hooks = %d begins / %d ends; want 2/2", len(obs.begins), len(obs.ends))
	}
	assertCallBeginMessages(t, obs)
	assertCallEstimateGrows(t, reg, obs)
	assertCallEndFlags(t, obs)
}

// assertCallBeginMessages pins the fused wire slice per call (call 1 = the user
// prompt; call 2 = prompt + assistant tool-call + tool result).
func assertCallBeginMessages(t *testing.T, obs *recordingObserver) {
	t.Helper()
	if obs.begins[0].idx != 0 || obs.begins[1].idx != 1 {
		t.Errorf("call indexes = %d,%d; want 0,1", obs.begins[0].idx, obs.begins[1].idx)
	}
	if len(obs.begins[0].msgs) != 1 || obs.begins[0].msgs[0].Role != "user" || obs.begins[0].msgs[0].Content != "read notes.txt" {
		t.Errorf("call 1 messages = %+v; want just the user prompt", obs.begins[0].msgs)
	}
	if len(obs.begins[1].msgs) != 3 {
		t.Errorf("call 2 messages = %d; want 3 (user + assistant tool-call + tool result)", len(obs.begins[1].msgs))
	}
}

// assertCallEstimateGrows pins that the per-call estimate over the fused messages
// grows across the turn (the CLI computes it via llm.EstimatePayload).
func assertCallEstimateGrows(t *testing.T, reg tools.Registry, obs *recordingObserver) {
	t.Helper()
	e1 := llm.EstimatePayload("persona", agentport.ToolDefs(reg), obs.begins[0].msgs)
	e2 := llm.EstimatePayload("persona", agentport.ToolDefs(reg), obs.begins[1].msgs)
	if e2 <= e1 {
		t.Errorf("the per-call estimate did not grow: e1=%d e2=%d", e1, e2)
	}
}

// assertCallEndFlags pins the final flag and the round reasons on each call-end.
func assertCallEndFlags(t *testing.T, obs *recordingObserver) {
	t.Helper()
	if obs.ends[0].final || len(obs.ends[0].reasons) != 1 || obs.ends[0].reasons[0] != "checking" {
		t.Errorf("call 1 end = %+v; want final=false, reasons=[checking]", obs.ends[0])
	}
	if !obs.ends[1].final || len(obs.ends[1].reasons) != 0 {
		t.Errorf("call 2 end = %+v; want final=true, no reasons", obs.ends[1])
	}
}

// TestAgentLoop_NoObserverIsSafe pins that the hook firings are inert without an
// observer (the seam's no-op default).
func TestAgentLoop_NoObserverIsSafe(t *testing.T) {
	gw := &fakeGateway{responses: []llm.Response{{Text: "hi"}}}
	a := &AgentLoop{Gateway: gw, Registry: tools.NewRegistry()}
	if _, err := a.Run(context.Background(), "q", nil); err != nil {
		t.Fatalf("Run without observer: %v", err)
	}
}

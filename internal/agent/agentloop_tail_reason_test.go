package agent

import (
	"context"
	"testing"

	"github.com/gosharplite/tellme/internal/domain/llm"
	"github.com/gosharplite/tellme/internal/domain/tools"
)

// Round-046 fold-review F-1(2): the tail-side half of the blank-reason policy.
// The loop's reasonsOf must hand OnCallEnd ONLY the reasons the renderer owner
// approved — a whitespace-only reason must never reach the tail. Before this pin
// a mutant that ignored `renders` at the tail filter leaked the blank reason into
// `roundReasons` and reddened NOTHING (not even the 60 s E2E), because the tail no
// longer re-checks each value (the deleted defensive guard — see the R-4
// precondition on OnCallEnd).
func TestTailReceivesOnlyRendererApprovedReasons(t *testing.T) {
	gw := &fakeGateway{responses: []llm.Response{
		{ToolCalls: []llm.ToolCall{
			{ID: "c1", Name: "read_files", Arguments: `{"filepaths":["a.txt"],"reason":"   "}`},
			{ID: "c2", Name: "read_files", Arguments: `{"filepaths":["b.txt"],"reason":"because"}`},
		}},
		{Text: "done"},
	}}
	obs := &recordingObserver{}
	a := &AgentLoop{
		Gateway:  gw,
		Registry: tools.NewRegistry(fakeTool{name: "read_files", result: "hi"}),
		Observer: obs,
		Lines:    fakeRenderer{},
	}
	if _, err := a.Run(context.Background(), "q", nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(obs.ends) == 0 {
		t.Fatal("no call-end observed")
	}
	got := obs.ends[0].reasons
	if len(got) != 1 || got[0] != "because" {
		t.Fatalf("tail reasons = %#v; want exactly [because] — the blank reason must be filtered by the owner (ui.ToolLineRenderer.ReasonLine)", got)
	}
}

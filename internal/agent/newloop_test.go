package agent

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	agentport "github.com/gosharplite/tellme/internal/domain/agent"
	"github.com/gosharplite/tellme/internal/domain/tools"
)

// samePtr reports whether a and b are the SAME pointer (used to pin field
// mapping by identity rather than by value, which would compare buffer/struct
// contents and could pass on distinct-but-equal objects).
func samePtr(a, b any) bool {
	va, vb := reflect.ValueOf(a), reflect.ValueOf(b)
	if va.Kind() != reflect.Pointer || vb.Kind() != reflect.Pointer {
		return false
	}
	return va.Pointer() == vb.Pointer()
}

// TestNewLoop_CopiesEverySpecField is the round-050 fold TD-1 parity guard.
//
// `agent.NewLoop` is the ENTIRE behaviour-preservation carrier of round 050 — a
// field-for-field copy of the loop's construction inputs — and `var _ Loop =
// (*AgentLoop)(nil)` proves only interface SATISFACTION, not that every spec
// field lands on the struct. Without this guard a future LoopSpec field the
// adapter forgets would pass the ZERO VALUE into the loop: RULE-C stays pure, the
// gate stays green, the construction site looks wired, and only an E2E path that
// exercises the new seam (if any) would notice. This test puts a DISTINCT
// sentinel in every LoopSpec field and asserts each landed on the built loop.
func TestNewLoop_CopiesEverySpecField(t *testing.T) {
	gw := &fakeGateway{}
	reg := tools.NewRegistry(fakeTool{name: "t"})
	sink := &recordingSink{}
	renderer := &fakeRenderer{}
	observer := &recordingObserver{}
	stderr := &bytes.Buffer{}
	clock := func() time.Time { return time.Date(2031, 1, 2, 3, 4, 5, 0, time.UTC) }

	spec := agentport.LoopSpec{
		Gateway:         gw,
		Registry:        reg,
		MaxLoops:        7,
		Stderr:          stderr,
		Now:             clock,
		EffectiveBudget: 1234,
		ToolUsage:       sink,
		Lines:           renderer,
		Observer:        observer,
	}

	loop, ok := NewLoop(spec).(*AgentLoop)
	if !ok {
		t.Fatalf("NewLoop did not return *AgentLoop")
	}

	if !samePtr(loop.Gateway, gw) {
		t.Errorf("Gateway not copied (got %v, want the sentinel)", loop.Gateway)
	}
	if !samePtr(loop.Registry, reg) {
		t.Errorf("Registry not copied")
	}
	if loop.MaxLoops != 7 {
		t.Errorf("MaxLoops = %d, want 7", loop.MaxLoops)
	}
	if !samePtr(loop.Stderr, stderr) {
		t.Errorf("Stderr not copied")
	}
	if loop.Now == nil || reflect.ValueOf(loop.Now).Pointer() != reflect.ValueOf(clock).Pointer() {
		t.Errorf("Now not copied")
	}
	if loop.EffectiveBudget != 1234 {
		t.Errorf("EffectiveBudget = %d, want 1234", loop.EffectiveBudget)
	}
	if !samePtr(loop.ToolUsage, sink) {
		t.Errorf("ToolUsage not copied")
	}
	if !samePtr(loop.Lines, renderer) {
		t.Errorf("Lines not copied")
	}
	if !samePtr(loop.Observer, observer) {
		t.Errorf("Observer not copied")
	}
}

package agent_test

// Round 050 (R5.4 of #92; ADR 0019) pins for the injected loop port
// (`Loop`/`LoopSpec`/`LoopFactory`). These prove the domain contract is the
// construction/execution seam a caller consumes, and that a LoopFactory builds a
// Loop from a LoopSpec without naming a concrete implementation. RULE-C purity of
// the contract is enforced by the layer-discipline gate (it imports stdlib +
// domain only), not re-asserted here.
//
// Scope (round-050 fold N-1): this test's value is the SHAPE/compile pin (the
// exported signatures + the factory round-trip); it exercises 2 of LoopSpec's 9
// fields by design. The full field-by-field MAPPING (LoopSpec -> the concrete
// loop) is owned by `internal/agent`'s TestNewLoop_CopiesEverySpecField (TD-1),
// not here.

import (
	"context"
	"testing"

	agentport "github.com/gosharplite/tellme/internal/domain/agent"
	"github.com/gosharplite/tellme/internal/domain/history"
)

// fakeLoop is a minimal agentport.Loop double driven by the spec it was built
// with, so a test can prove a LoopFactory receives (and may observe) the spec.
type fakeLoop struct {
	spec agentport.LoopSpec
	got  struct {
		prompt string
		prior  []history.Entry
	}
}

func (f *fakeLoop) Run(_ context.Context, prompt string, prior []history.Entry) (agentport.Result, error) {
	f.got.prompt = prompt
	f.got.prior = prior
	return agentport.Result{Answer: "ok"}, nil
}

func TestLoopFactory_BuildsLoopFromSpec(t *testing.T) {
	var factory agentport.LoopFactory = func(spec agentport.LoopSpec) agentport.Loop {
		return &fakeLoop{spec: spec}
	}

	spec := agentport.LoopSpec{MaxLoops: 7, EffectiveBudget: 1234}
	loop := factory(spec)

	fl, ok := loop.(*fakeLoop)
	if !ok {
		t.Fatalf("factory did not return the constructed loop")
	}
	if fl.spec.MaxLoops != 7 || fl.spec.EffectiveBudget != 1234 {
		t.Errorf("spec not carried to the loop: %+v", fl.spec)
	}

	res, err := loop.Run(context.Background(), "ping", []history.Entry{{Prompt: "q", Answer: "a"}})
	if err != nil {
		t.Fatalf("Run err = %v, want nil", err)
	}
	if res.Answer != "ok" {
		t.Errorf("Result.Answer = %q, want ok", res.Answer)
	}
	if fl.got.prompt != "ping" || len(fl.got.prior) != 1 {
		t.Errorf("Run did not receive the prompt+prior: prompt=%q prior=%+v", fl.got.prompt, fl.got.prior)
	}
}

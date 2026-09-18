package agent

import (
	agentport "github.com/gosharplite/tellme/internal/domain/agent"
)

// NewLoop is the concrete adapter for the domain agentport.Loop port (round 050;
// R5.4 of #92; ADR 0019). It builds an *AgentLoop from the domain-typed
// LoopSpec, so the application layer obtains the loop through the injected
// LoopFactory without naming this package's types.
//
// The assignment is a pure field-for-field copy of the loop's construction
// inputs (the same 9 fields the CLI used to set on the literal), so the loop's
// behaviour is unchanged.
func NewLoop(spec agentport.LoopSpec) agentport.Loop {
	return &AgentLoop{
		Gateway:         spec.Gateway,
		Registry:        spec.Registry,
		MaxLoops:        spec.MaxLoops,
		Stderr:          spec.Stderr,
		Now:             spec.Now,
		EffectiveBudget: spec.EffectiveBudget,
		ToolUsage:       spec.ToolUsage,
		Lines:           spec.Lines,
		Observer:        spec.Observer,
	}
}

// compile-time proof that the concrete loop satisfies the domain port.
var _ agentport.Loop = (*AgentLoop)(nil)

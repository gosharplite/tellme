package cli

import (
	agentport "github.com/gosharplite/tellme/internal/domain/agent"
	"github.com/gosharplite/tellme/internal/domain/llm"
)

// compositeObserver composes the round-034 per-call observer (the decomposed
// tool-call block renderer + the per-call frame/tail) with the round-019
// spinner into one agentport.LoopObserver, so the loop keeps a single observer
// seam (ADR 0005 D1). T001 landing skeleton: the seam + nil-safe delegation;
// the block renderer is bound on the prompt path in round 034 T025/T027.
//
// The call-begin/call-end hooks go to `call`; every waiting-phase hook goes to
// `spinner`. A nil dependency is a no-op, so the composite is safe to build even
// when the spinner is gated off (round-019 FR-006).
type compositeObserver struct {
	call    agentport.CallObserver
	spinner agentport.LoopObserver
}

// OnCallBegin forwards the call-begin hook to the call observer.
func (c compositeObserver) OnCallBegin(callIndex int, messages []llm.Message) {
	if c.call != nil {
		c.call.OnCallBegin(callIndex, messages)
	}
}

// OnCallEnd forwards the call-end hook to the call observer.
func (c compositeObserver) OnCallEnd(callIndex int, usage llm.Usage, roundReasons []string, final bool) {
	if c.call != nil {
		c.call.OnCallEnd(callIndex, usage, roundReasons, final)
	}
}

// OnInferenceStart forwards the model-wait phase to the spinner.
func (c compositeObserver) OnInferenceStart() {
	if c.spinner != nil {
		c.spinner.OnInferenceStart()
	}
}

// OnInferenceEnd forwards the model-return phase to the spinner.
func (c compositeObserver) OnInferenceEnd() {
	if c.spinner != nil {
		c.spinner.OnInferenceEnd()
	}
}

// OnToolsStart forwards the tool-execution phase to the spinner.
func (c compositeObserver) OnToolsStart(names []string) {
	if c.spinner != nil {
		c.spinner.OnToolsStart(names)
	}
}

// OnToolsEnd forwards the tool-batch end to the spinner.
func (c compositeObserver) OnToolsEnd() {
	if c.spinner != nil {
		c.spinner.OnToolsEnd()
	}
}

// BeforeToolLog yields the line to a tool-loop log write via the spinner.
func (c compositeObserver) BeforeToolLog() {
	if c.spinner != nil {
		c.spinner.BeforeToolLog()
	}
}

// AfterToolLog restores the spinner after a tool-loop log write.
func (c compositeObserver) AfterToolLog() {
	if c.spinner != nil {
		c.spinner.AfterToolLog()
	}
}

// Compile-time port conformance (round 034 ADR 0005 D1).
var _ agentport.LoopObserver = compositeObserver{}

// noopCallObserver is the default call observer: it renders nothing (the seam's
// no-op default, ADR 0005 D1). The real block renderer is bound in T025/T027.
type noopCallObserver struct{}

// OnCallBegin is a no-op.
func (noopCallObserver) OnCallBegin(int, []llm.Message) {}

// OnCallEnd is a no-op.
func (noopCallObserver) OnCallEnd(int, llm.Usage, []string, bool) {}

// Compile-time port conformance.
var _ agentport.CallObserver = noopCallObserver{}

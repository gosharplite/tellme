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

// OnCallEnd forwards the call-end hook to the call observer, yielding the
// progress indicator first on a non-final call.
//
// Round 035 (issue #72): a PHASE-BOUNDARY yield. The round-034 per-call tail is
// written by the `call` half while the round-019/025 spinner is still live (the
// loop's Spinner.OnToolsEnd is a no-op), so the tail's first line shares the
// spinner's frame row and the frame residue survives the finished turn. The
// indicator is synchronously CLEARED before the tail write and NOT resumed — at
// this boundary the next waiting phase (the next call's OnInferenceStart) or the
// turn's Stop() re-activates it. A clear+resume would only relocate the residue
// (the next call's frame write opens with a bare "\n", stranding the resumed
// frame on the row above). The FINAL call's tail is deferred past the answer
// (ADR 0005 G5), so nothing is written at its call-end and it must not touch the
// spinner — keeping the final path byte-identical.
func (c compositeObserver) OnCallEnd(callIndex int, usage llm.Usage, roundReasons []string, final bool) {
	if !final {
		c.yieldIndicatorBeforeTail()
	}
	if c.call != nil {
		c.call.OnCallEnd(callIndex, usage, roundReasons, final)
	}
}

// yieldIndicatorBeforeTail synchronously yields the progress indicator so a
// non-final tail write starts on its own cleared line (round 035). It is
// nil-safe (a gated-off spinner is a no-op). It deliberately does NOT restore —
// see OnCallEnd for why (the phase boundary owns re-activation). Round 045: the
// tail yield uses the same YieldIndicator vocabulary the loop uses, so no
// tool-log-named hook is used for this non-log yield.
func (c compositeObserver) yieldIndicatorBeforeTail() {
	if c.spinner != nil {
		c.spinner.YieldIndicator()
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

// YieldIndicator forwards the loop's yield hook (clear-only) to the spinner
// half. Round 045: this is the single vocabulary for a yield — the loop's direct
// log writes and this composite's per-call tail both reach the spinner through
// it (the yield POLICY lives with ui.YieldController; ADR 0014).
func (c compositeObserver) YieldIndicator() {
	if c.spinner != nil {
		c.spinner.YieldIndicator()
	}
}

// RestoreIndicator forwards the loop's restore hook (resume) to the spinner half.
func (c compositeObserver) RestoreIndicator() {
	if c.spinner != nil {
		c.spinner.RestoreIndicator()
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

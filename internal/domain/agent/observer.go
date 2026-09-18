// Package agent defines the agent-loop observation port (round-019 research
// Decision 7). The loop invokes these hooks around its Complete calls and its
// diagnostic-stream writes, so a CLI-injected presenter (the spinner) can label,
// clear, and restore per waiting phase. The port is network-free and knows
// nothing about presentation.
//
// Round 045 (R3 of #92) splits the port's overloaded yield pair: BeforeToolLog /
// AfterToolLog conflated "a tool-log line is about to be written" with "the
// indicator must yield", yet one of their four call sites (the round-035 per-call
// tail) is NOT a tool-log write. They are replaced by the intent-named
// YieldIndicator (clear-only) / RestoreIndicator (resume) pair. The port states
// the loop's ROUTE to a yield; the yield POLICY — when a clear is permitted, that
// a clear never resumes, and that restoration is a phase-boundary act — is owned
// by internal/ui (ui.YieldController; ADR 0014).
package agent

// LoopObserver observes one prompt run's waiting phases.
//
// Lifecycle (round-019 research Decision 7): the loop calls OnInferenceStart
// before each Gateway.Complete and OnInferenceEnd after it; OnToolsStart /
// OnToolsEnd around a tool batch; and YieldIndicator / RestoreIndicator around
// each diagnostic write it performs directly, so an implementer can clear the
// indicator for the write and (when the phase boundary calls for it) restore it
// afterwards.
type LoopObserver interface {
	// CallObserver adds the round-034 per-AI-endpoint-call hooks (call-begin /
	// call-end; ADR 0005 D1) to the same port, so the loop keeps a single
	// observer seam. The CLI's compositeObserver implements both facets.
	CallObserver
	// OnInferenceStart signals that the run has begun awaiting the model.
	OnInferenceStart()
	// OnInferenceEnd signals that the model's response has arrived.
	OnInferenceEnd()
	// OnToolsStart signals that the run has begun executing the named tools.
	OnToolsStart(names []string)
	// OnToolsEnd signals that the tool batch has finished.
	OnToolsEnd()
	// YieldIndicator signals that the loop is about to write directly to the
	// diagnostic stream, so the progress indicator should YIELD (clear). It is
	// clear-only: the loop does NOT resume here — restoration is a phase-boundary
	// act owned by the presenter (round 045 / ADR 0014).
	YieldIndicator()
	// RestoreIndicator signals that a yield is over and the indicator may resume.
	// Whether a resume is appropriate at this point is the presenter's policy
	// (ui.YieldController), not the loop's.
	RestoreIndicator()
}

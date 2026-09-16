// Package agent defines the agent-loop observation port (round-019 research
// Decision 7). The loop invokes these hooks around its Complete calls and its
// logStep stderr write, so a CLI-injected presenter (the spinner) can label,
// clear, and restore per waiting phase. The port is network-free and knows
// nothing about presentation.
package agent

// LoopObserver observes one prompt run's waiting phases.
//
// Lifecycle (round-019 research Decision 7): the loop calls OnInferenceStart
// before each Gateway.Complete and OnInferenceEnd after it; OnToolsStart /
// OnToolsEnd around a tool batch; and BeforeToolLog / AfterToolLog around each
// logStep stderr write, so an implementer can clear the indicator for the log
// line and restore it afterwards.
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
	// BeforeToolLog signals that a tool-loop log line is about to be written to
	// the diagnostic stream (the observer should yield the line).
	BeforeToolLog()
	// AfterToolLog signals that the tool-loop log line has been written.
	AfterToolLog()
}

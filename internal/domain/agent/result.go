package agent

import (
	"github.com/gosharplite/tellme/internal/domain/history"
	"github.com/gosharplite/tellme/internal/domain/llm"
	"github.com/gosharplite/tellme/internal/domain/tools"
)

// This file carries the loop's crossing CONTRACTS — the types and the pure
// projection that BOTH the loop implementation (`internal/agent`) and its CLI
// caller (`internal/cli`) consume. They were relocated here, verbatim, in
// round 049 (R5.3 of #92; ADR 0018) so that:
//
//   - the contracts have exactly ONE home, at the domain tier (RULE-C-pure:
//     this file imports only domain packages); and
//   - the surviving `internal/cli -> internal/agent` coupling is reduced to the
//     single `AgentLoop` construction, making the later sub-slice 2 inversion
//     (into an injected domain port) provably edge-sized.
//
// Deliberately NO alias is re-exported from `internal/agent` (round-049
// clarify Q3 → (a): one concept → one name → one home).

// Result is the outcome of one prompt run: the final answer text, the ordered
// tool steps performed (for persistence), and the provider's reported usage of
// the final completion. Surfacing Usage here — rather than discarding it inside
// the loop — is what lets the CLI render the post-turn payload status line
// (round-009 BLOCKER-2). On a multi-step tool run it is the usage of the final
// completion (the response that produced the answer).
//
// Formerly `internal/agent.AgentResult` (relocated round 049, ADR 0018).
type Result struct {
	Answer string
	Steps  []history.Step
	Usage  llm.Usage
	// Calls holds EVERY provider call's usage for the turn, in call order
	// (round 018), so the CLI can compute the turn cost (`$#2`) and persist each
	// call to the usage log. `Usage` remains the just-returned (final) call.
	Calls []llm.Usage
}

// ErrIncomplete reports that a tool-using run could not reach a final answer —
// the iteration bound was reached, or the model requested a tool that tellme
// does not provide. The CLI maps it to the frozen class phrase
// `the tool request failed` and exit code 7 (round-008 FR-010).
//
// Formerly `internal/agent.ErrIncomplete` (relocated round 049, ADR 0018).
type ErrIncomplete struct {
	Reason string
	Err    error
}

// Error renders the reason plus the underlying cause, if any.
func (e *ErrIncomplete) Error() string {
	if e.Err != nil {
		return e.Reason + ": " + e.Err.Error()
	}
	return e.Reason
}

// Unwrap exposes the underlying cause for errors.Is / errors.As.
func (e *ErrIncomplete) Unwrap() error { return e.Err }

// ToolDefs projects a tool registry into the wire definitions offered to the
// model. Exported so the CLI's pre-flight estimate counts exactly what the loop
// sends (round-011 RF-1), mirroring the BuildMessages reuse (round 009).
//
// Formerly `internal/agent.ToolDefs` (relocated round 049, ADR 0018).
func ToolDefs(reg tools.Registry) []llm.ToolDef {
	if reg == nil {
		return nil
	}
	ts := reg.Tools()
	defs := make([]llm.ToolDef, 0, len(ts))
	for _, t := range ts {
		defs = append(defs, llm.ToolDef{Name: t.Name(), Description: t.Description(), Parameters: t.Parameters()})
	}
	return defs
}

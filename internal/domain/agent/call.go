package agent

import "github.com/gosharplite/tellme/internal/domain/llm"

// CallObserver observes each AI-endpoint call of one prompt run (ADR 0005 D1):
// a call-begin hook (fired right after the request is assembled) and a
// call-end hook (fired at the call's return). The CLI composite presenter
// implements it; the loop emits nothing itself.
//
//	OnCallBegin(callIndex, messages)                          // CLI computes llm.EstimatePayload(person, toolDefs, messages)
//	OnCallEnd(callIndex, usage, roundReasons, final)          // usage.PromptTokens IS the measured payload
//
// `messages` is the loop's FUSED base+turn wire slice (not llm.Request, so the
// adapter's prompt-vs-messages rule is not duplicated — ADR 0005 D2).
type CallObserver interface {
	// OnCallBegin signals that an AI-endpoint call is about to be made. callIndex
	// is the 0-based call index within the turn; messages is the fused base+turn
	// slice the request will carry.
	OnCallBegin(callIndex int, messages []llm.Message)
	// OnCallEnd signals that an AI-endpoint call has returned. roundReasons are
	// the reasons of the calls executed in the round that just finished (empty on
	// a final answer); final is true for the call that produced the answer.
	OnCallEnd(callIndex int, usage llm.Usage, roundReasons []string, final bool)
}

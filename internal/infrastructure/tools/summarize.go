package tools

import (
	"context"
	"encoding/json"
	"errors"

	domhistory "github.com/gosharplite/tellme/internal/domain/history"
	dllm "github.com/gosharplite/tellme/internal/domain/llm"
	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
)

var errSummarizeNotImplemented = errors.New("summarize_history: not implemented")

// summarizeHistory is the LLM-backed session-summarisation tool (round-008
// research Decision 10): it reads the persisted conversation through the
// injected Store and requests a summary through the injected Gateway, returning
// the summary as the tool result WITHOUT mutating stored records (FR-013). Its
// parameter schema is empty; it runs inside the same bounded loop.
type summarizeHistory struct {
	store   domhistory.Store
	gateway dllm.Gateway
}

// Name is the wire-valid canonical identifier (round-008 BLOCKER-1).
func (*summarizeHistory) Name() string { return "summarize_history" }

// Description is the model-facing summary.
func (*summarizeHistory) Description() string { return "Summarise the conversation so far." }

// Parameters is the empty JSON-schema for the tool's arguments.
func (*summarizeHistory) Parameters() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{}}`)
}

// Execute is the Foundational skeleton (round 008 T002): the LLM-backed body is
// implemented in a Feature phase.
func (*summarizeHistory) Execute(_ context.Context, _ string) (string, error) {
	return "", errSummarizeNotImplemented
}

// NewSummarizeHistoryTool returns the LLM-backed session-summarisation tool,
// injected with the session-history store and the provider gateway.
func NewSummarizeHistoryTool(store domhistory.Store, gateway dllm.Gateway) domaintools.Tool {
	return &summarizeHistory{store: store, gateway: gateway}
}

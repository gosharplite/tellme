package tools

import (
	"context"
	"encoding/json"
	"fmt"

	domhistory "github.com/gosharplite/tellme/internal/domain/history"
	dllm "github.com/gosharplite/tellme/internal/domain/llm"
	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
)

// summarizePrompt is the instruction sent alongside the persisted conversation
// when the summarise tool asks the model for a condensed summary.
const summarizePrompt = "Summarise the conversation so far in a few sentences."

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

// Execute reads the persisted conversation, asks the injected gateway to
// summarise it, and returns the summary text. It only READS the store — it never
// appends or rewrites any stored record (FR-013), so earlier conversation
// records are left byte-unchanged.
func (t *summarizeHistory) Execute(ctx context.Context, _ string) (string, error) {
	entries, err := t.store.Load()
	if err != nil {
		return "", fmt.Errorf("summarize_history: %w", err)
	}
	prior := make([]dllm.Message, 0, len(entries)*2)
	for _, e := range entries {
		prior = append(prior, dllm.Message{Role: "user", Content: e.Prompt})
		prior = append(prior, dllm.Message{Role: "assistant", Content: e.Answer})
	}
	resp, err := t.gateway.Complete(ctx, dllm.Request{Prompt: summarizePrompt, Messages: prior})
	if err != nil {
		return "", fmt.Errorf("summarize_history: %w", err)
	}
	return resp.Text, nil
}

// NewSummarizeHistoryTool returns the LLM-backed session-summarisation tool,
// injected with the session-history store and the provider gateway.
func NewSummarizeHistoryTool(store domhistory.Store, gateway dllm.Gateway) domaintools.Tool {
	return &summarizeHistory{store: store, gateway: gateway}
}

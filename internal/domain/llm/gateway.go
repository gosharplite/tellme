// Package llm defines the provider-agnostic reasoning gateway port: the
// domain-facing abstraction the CLI calls to complete one prompt. It knows
// nothing about the wire protocol or transport (round-004 research Decision 1),
// so the transport is swappable and the turn is fake-testable.
package llm

import (
	"context"
	"encoding/json"
)

// ToolDef is a tool definition offered to the model on a request (round-008
// research Decision 2). The adapter sends the wire-valid snake_case Name, the
// Description, and the Parameters JSON-schema in the top-level `tools` array.
type ToolDef struct {
	Name        string
	Description string
	Parameters  json.RawMessage
}

// ToolCall is a model's structured request to run a tool (round-008 research
// Decision 2): the wire call id, the tool name, and the raw arguments string.
type ToolCall struct {
	ID        string
	Name      string
	Arguments string
	// Signature carries a provider-specific opaque token that must be echoed
	// back verbatim when the call is replayed on a later request — the Vertex AI
	// Gemini 3 `thoughtSignature`. Empty for providers that do not use one
	// (e.g. the OpenAI family).
	Signature string
}

// Message is one prior conversation message carried on a request: the role
// ("user", "assistant", or "tool") and its content. For a tool turn it also
// carries the assistant's tool-call requests (ToolCalls) or a tool result's
// ToolCallID (round-008 research Decision 2). It is empty on the request's first
// turn; on later turns it carries the persisted conversation (round-007
// research Decision 2 / RF-3).
// Media, when non-empty, carries image (or other media) content attached to
// this message (round 062; ADR 0032). A message with NO media serializes
// exactly as before (a plain string `content`), so the text path is
// byte-identical; a message WITH media is serialized by the adapter as a
// content array (a text part when the message states text, then one image part
// per MediaPart). Media rides a `user` message (the ref-less capability the
// OpenAI-compatible family accepts inline).
type Message struct {
	Role       string
	Content    string
	Media      []MediaPart
	ToolCalls  []ToolCall
	ToolCallID string
}

// Request is a single provider completion request. Prompt is the current turn's
// prompt; Messages is the resumed conversation that precedes it (empty for a
// fresh conversation); Tools are the tool definitions offered to the model
// (empty when no tools are registered — the payload is then byte-identical to
// rounds 004–007). The adapter sends Messages followed by the current prompt.
type Request struct {
	Prompt   string
	Messages []Message
	Tools    []ToolDef
}

// Usage is the provider's reported token usage for a completion (round-009
// research Decision 2). Reported is false when the provider response carried no
// usage block, so the caller omits the post-turn payload status line.
//
// Round 018 widens it with the call's cached (H) and reasoning (Th) token counts
// for the post-turn metrics line. `CompletionTokens` is the EXCLUSIVE completion
// count — the OpenAI-compatible wire `completion_tokens` MINUS `reasoning_tokens`
// — so `C`/`Th` are disjoint and additive (`O = C + Th` never double-counts)
// (round-018 research Decision 1 / FR-002).
type Usage struct {
	Reported         bool
	PromptTokens     int
	CachedTokens     int
	CompletionTokens int
	ThinkingTokens   int
	TotalTokens      int
}

// Response is the normalized answer extracted from a provider response: the
// answer text (empty when the model only requested tools), any structured
// tool-call requests, and — round 009 — the provider's reported Usage.
type Response struct {
	Text      string
	ToolCalls []ToolCall
	Usage     Usage
}

// ProviderError is the single typed error for a provider or transport failure
// (round-004 research Decision 5). The CLI maps it to the frozen class phrase
// `the provider request failed` and exit code 6.
type ProviderError struct {
	Provider string
	Err      error
}

// Error renders the provider name plus the underlying cause.
func (e *ProviderError) Error() string {
	if e.Provider != "" {
		return "provider " + e.Provider + ": " + e.Err.Error()
	}
	return e.Err.Error()
}

// Unwrap exposes the underlying cause for errors.Is / errors.As.
func (e *ProviderError) Unwrap() error { return e.Err }

// Gateway is the provider-agnostic completion port.
type Gateway interface {
	Complete(ctx context.Context, req Request) (Response, error)
}

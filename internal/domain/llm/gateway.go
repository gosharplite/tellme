// Package llm defines the provider-agnostic reasoning gateway port: the
// domain-facing abstraction the CLI calls to complete one prompt. It knows
// nothing about the wire protocol or transport (round-004 research Decision 1),
// so the transport is swappable and the turn is fake-testable.
package llm

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/gosharplite/tellme/internal/domain/tools"
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
	Media      []tools.MediaPart
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
//
// Round 078 (ADR 0050) adds the typed failure FACTS the retry predicate needs —
// the adapters set them where the reason is known, so retryability is classified
// without parsing the message (a `Status int` and a `Transport bool`):
//   - Transport is true when the failure came from the CONNECTION (a dial
//     failure, EOF / connection reset / broken pipe / GOAWAY / TLS failure / a
//     request timeout) rather than from the status/body/decode.
//   - Status is the HTTP status code when the failure WAS an HTTP-status
//     failure (0 otherwise — a transport failure, a decode error, a truncation).
type ProviderError struct {
	Provider string
	Err      error
	// Status is the provider's HTTP status code, or 0 when the failure was not
	// an HTTP-status failure (round 078).
	Status int
	// Transport marks a connection-level failure (round 078).
	Transport bool
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

// retryableStatus reports whether an HTTP status is transient: 429 (rate
// limited) or any 5xx (a server-side failure). Every other status — 4xx
// (bad request / auth / not found / …) and 0 (no status) — is not.
func retryableStatus(status int) bool {
	return status == 429 || (status >= 500 && status <= 599)
}

// Retryable reports whether a provider failure is worth retrying (round 078;
// ADR 0050 D2). It is the SINGLE domain owner of the retryability policy — the
// CLI's retry decorator consults it and never inspects the error text (NFR-003).
//
// A failure is retryable iff it is a *ProviderError whose typed facts mark it
// transient: a TRANSPORT failure (connection level) or an HTTP 429/5xx status.
// A non-*ProviderError, a decode error, or the round-030 output-cap truncation
// (neither flag set) is NOT retryable.
func Retryable(err error) bool {
	var pe *ProviderError
	if !errors.As(err, &pe) {
		return false
	}
	if pe.Transport {
		return true
	}
	return retryableStatus(pe.Status)
}

// Gateway is the provider-agnostic completion port.
type Gateway interface {
	Complete(ctx context.Context, req Request) (Response, error)
}

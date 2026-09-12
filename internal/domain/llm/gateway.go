// Package llm defines the provider-agnostic reasoning gateway port: the
// domain-facing abstraction the CLI calls to complete one prompt. It knows
// nothing about the wire protocol or transport (round-004 research Decision 1),
// so the transport is swappable and the turn is fake-testable.
package llm

import "context"

// Message is one prior conversation message carried on a request: the role
// ("user" or "assistant") and its content. It is empty on the request's first
// turn; on later turns it carries the persisted conversation (round-007 research
// Decision 2 / RF-3).
type Message struct {
	Role    string
	Content string
}

// Request is a single provider completion request. Prompt is the current turn's
// prompt; Messages is the resumed conversation that precedes it (empty for a
// fresh conversation). The adapter sends Messages followed by the current prompt.
type Request struct {
	Prompt   string
	Messages []Message
}

// Response is the normalized answer extracted from a provider response.
type Response struct {
	Text string
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

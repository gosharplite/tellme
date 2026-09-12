// Package llm defines the provider-agnostic reasoning gateway port: the
// domain-facing abstraction the CLI calls to complete one prompt. It knows
// nothing about the wire protocol or transport (round-004 research Decision 1),
// so the transport is swappable and the turn is fake-testable.
package llm

import "context"

// Request is a single provider completion request. It carries only the
// operator's prompt; the provider-specific endpoint, credential, model, and
// limits are bound to the concrete adapter at construction time.
type Request struct {
	Prompt string
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

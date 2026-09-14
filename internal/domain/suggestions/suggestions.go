// Package suggestions defines tellme's domain port for prompt suggestions
// (round 015). It is the network-free, dependency-light seam the interactive
// prompt draws from: a multi-source engine (shared log + session + workspace +
// registered tools) sits behind it as the adapter/coordinator
// (internal/app/suggestions), so the prompt stays testable with a fake source.
package suggestions

import "context"

// Suggestion is one candidate completion offered by the suggestion engine: a
// recent prompt (from the shared global log or the active session), a workspace
// path, or a registered tool name (spec.md Key Entities).
type Suggestion struct {
	// Text is the operator-facing suggestion text.
	Text string
}

// Service is the domain port for the multi-source suggestion engine. Given a
// query it returns a deduplicated, capped list of suggestions (round-015
// research Decision 2). Implemented by internal/app/suggestions; consumed by the
// interactive prompt (internal/ui/tui/prompt).
type Service interface {
	// Suggest returns up to a fixed cap of candidates matching query. An empty
	// query yields the newest recent prompts. Implementations must honour ctx for
	// cooperative debounce cancellation and must skip noisy directories
	// (round-015 PR #38 review directive ③).
	Suggest(ctx context.Context, query string) []Suggestion

	// Close drains any background work (round-015 research Decision 3 /
	// PR #38 review directive ⑤) before the process exits.
	Close(ctx context.Context) error
}

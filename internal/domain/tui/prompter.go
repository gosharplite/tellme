// Package tui declares the domain port through which the application tier
// reaches the interactive TUI prompt, without importing the presentation
// package that implements it.
//
// Round 048 (R5.2 of #92; ADR 0017) introduced this port to remove the residual
// layer-discipline edge `internal/cli -> internal/ui/tui/prompt` (RULE-E,
// ADR 0016): the CLI now depends on this domain contract, the presentation tier
// satisfies it (an internal/ui adapter), and the composition root wires them.
//
// The port references only stdlib (context/io/time) — it is RULE-C-pure.
package tui

import (
	"context"
	"io"
	"time"
)

// Source yields the candidate suggestions for the current query. It carries ctx
// so a superseded fetch can be cancelled. It mirrors internal/ui/tui/prompt's
// Source interface verbatim, so a Source value is assignable to it and vice
// versa (Go structural interface satisfaction).
type Source interface {
	Suggest(ctx context.Context, query string) []string
}

// Prompter runs the interactive terminal prompt for one invocation: it returns
// the composed prompt text plus whether a prompt was submitted (ok). It is the
// injection seam that decouples internal/cli from the Bubble Tea presentation
// package (ADR 0017). An implementation MUST NOT write to standard output — the
// composed text is returned to the caller, which owns the diagnostic stream.
type Prompter interface {
	// Run drives the prompt bound to the injected streams and returns the
	// composed text plus whether it was submitted. debounce is the
	// suggestion-refresh debounce (<=0 refreshes synchronously).
	Run(ctx context.Context, in io.Reader, out io.Writer, src Source, debounce time.Duration) (string, bool, error)

	// DefaultDebounceDuration is the suggestion-refresh debounce default applied
	// when no override is configured.
	DefaultDebounceDuration() time.Duration
}

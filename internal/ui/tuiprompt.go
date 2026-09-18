package ui

import (
	"context"
	"io"
	"time"

	domaintui "github.com/gosharplite/tellme/internal/domain/tui"
	"github.com/gosharplite/tellme/internal/ui/tui/prompt"
)

// TUIPrompter is the presentation-tier adapter that satisfies the domain
// domaintui.Prompter port (round 048 / ADR 0017): it delegates the interactive
// prompt's Run call and debounce default to internal/ui/tui/prompt.
//
// It lives in internal/ui (tier 5) because RULE-A (ADR 0011 D1) forbids a lower
// tier from importing internal/ui/tui/prompt (also tier 5); the composition root
// (cmd/tellme) injects it into internal/cli, so internal/cli no longer imports
// the TUI package (removing the RULE-E residual edge).
type TUIPrompter struct{}

// Run drives the interactive prompt bound to the injected streams (see
// prompt.Run). The Source value is passed through untouched — domaintui.Source
// and prompt.Source have identical method sets.
func (TUIPrompter) Run(ctx context.Context, in io.Reader, out io.Writer, src domaintui.Source, debounce time.Duration) (string, bool, error) {
	return prompt.Run(ctx, in, out, src, debounce)
}

// DefaultDebounceDuration is the suggestion-refresh debounce default (see
// prompt.DefaultDebounceDuration).
func (TUIPrompter) DefaultDebounceDuration() time.Duration { return prompt.DefaultDebounceDuration }

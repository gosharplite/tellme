package ui

import (
	"context"
	"io"
	"testing"
	"time"

	domaintui "github.com/gosharplite/tellme/internal/domain/tui"
	"github.com/gosharplite/tellme/internal/ui/tui/prompt"
)

// Compile-time contract: the presentation adapter satisfies the domain port
// (round 048 / ADR 0017).
var _ domaintui.Prompter = TUIPrompter{}

// TestTUIPrompterSatisfiesPort: the adapter's debounce default mirrors the TUI
// package it delegates to.
func TestTUIPrompterSatisfiesPort(t *testing.T) {
	if got := (TUIPrompter{}).DefaultDebounceDuration(); got != prompt.DefaultDebounceDuration {
		t.Fatalf("DefaultDebounceDuration = %v, want %v", got, prompt.DefaultDebounceDuration)
	}
}

// sharedSource satisfies both domaintui.Source and prompt.Source — the assertions
// pin that one concrete type's method set covers both (the load-bearing
// assignability the adapter relies on is pinned by TUIPrompter.Run's body
// compiling below).
type sharedSource struct{}

func (sharedSource) Suggest(_ context.Context, _ string) []string { return nil }

var (
	_ domaintui.Source = sharedSource{}
	_ prompt.Source    = sharedSource{}
	// The load-bearing assignability: a domaintui.Source is directly usable where
	// prompt.Source is expected (identical method sets), so the adapter needs no
	// wrapper. A future method added to prompt.Source breaks this at compile time.
	_ prompt.Source = domaintui.Source(nil)
)

// TUIPrompter.Run must keep this exact signature (the port contract) — a drift
// becomes a compile error. The Bubble Tea runtime itself is exercised end to end
// by the E2E TUI suite (not here, to avoid a real TTY).
var _ func(context.Context, io.Reader, io.Writer, domaintui.Source, time.Duration) (string, bool, error) = TUIPrompter{}.Run

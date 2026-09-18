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

// sharedSource has the method set both domaintui.Source and prompt.Source
// require; the two assertions pin the pass-through (a domaintui.Source is
// assignable to prompt.Source) at compile time.
type sharedSource struct{}

func (sharedSource) Suggest(_ context.Context, _ string) []string { return nil }

var (
	_ domaintui.Source = sharedSource{}
	_ prompt.Source    = sharedSource{}
)

// TUIPrompter.Run must keep this exact signature (the port contract) — a drift
// becomes a compile error. The Bubble Tea runtime itself is exercised end to end
// by the E2E TUI suite (not here, to avoid a real TTY).
var _ func(context.Context, io.Reader, io.Writer, domaintui.Source, time.Duration) (string, bool, error) = TUIPrompter{}.Run

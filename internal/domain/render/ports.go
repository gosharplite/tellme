// Package render declares the CLI's presentation seams (round 051; R5.5 of
// #92; ADR 0020). It is the domain-owned contract through which internal/cli
// obtains its status/tail lines, its answer renderer, and its turn progress
// indicator — so the CLI names no internal/ui type. The BYTES (sanitize / fold /
// cap / format policy) stay single-owned in internal/ui, which implements these
// ports; the domain owns only the shapes. RULE-C-pure: stdlib + domain types
// only.
package render

import (
	"io"
	"time"

	agentport "github.com/gosharplite/tellme/internal/domain/agent"
	"github.com/gosharplite/tellme/internal/domain/history"
	"github.com/gosharplite/tellme/internal/domain/metrics"
	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
)

// Lines renders the CLI's status/tail diagnostic lines. It mirrors the
// internal/ui formatters; the adapter delegates to them verbatim.
type Lines interface {
	InputCaptured(t time.Time) string
	TurnOpening(turn int, mode string) string
	TurnGap() string
	// PayloadStatus renders the MEASURED payload line (`<tokens>/<budget> tokens
	// - <mode> - <model>`; round 009, unchanged by round 057).
	PayloadStatus(t time.Time, tokens, budget int, mode, model string, estimated bool) string
	// PayloadEstimate renders the ESTIMATED pre-flight payload line (round 057;
	// ADR 0027): `+<delta> ~<tokens> tokens - <mode> - <model>`. The delta is the
	// increment over the previous estimate (`0` when there is none) and the budget
	// is not shown.
	PayloadEstimate(t time.Time, tokens, delta int, mode, model string) string
	Metrics(t time.Time, provider string, u metrics.UsageCounts) string
	Ready(lastCallCost, turnCost, sessionCost float64, sessionMiss, sessionHit, sessionOut int, hitRate float64) string
	ToolReason(t time.Time, reason string) string
	ToolUsage(rows []history.ToolUsageRow) string
	// DefaultToolOutputIdleGap is the `[Tool Output]` block's idle-gap default
	// (ADR 0009 D3).
	DefaultToolOutputIdleGap() time.Duration
}

// Answer renders a Markdown answer to ANSI (round 006). It mirrors the CLI's
// former in-package `answerRenderer` interface.
type Answer interface {
	// Render returns the rendered answer and whether rendering degraded (in which
	// case the returned string is the sanitized raw fallback).
	Render(markdown string, width int) (text string, degraded bool)
	// WarnDegraded emits the one-time degradation warning.
	WarnDegraded(w io.Writer)
}

// Indicator is the turn progress indicator (the spinner) as the CLI drives it.
// It is the loop's waiting-phase observer (agentport.LoopObserver) plus a Stop
// that clears it for the rest of the run.
type Indicator interface {
	agentport.LoopObserver
	Stop()
}

// TurnProgress bundles the coupled turn-scoped presentation objects a turn
// builds together: the progress Indicator (nil when the spinner is gated off)
// and the `[Tool Output]` sink (always present; it renders unconditionally).
type TurnProgress struct {
	Indicator  Indicator
	ToolOutput domaintools.OutputSink
}

// ProgressFactory builds a turn's TurnProgress over the diagnostic stream, the
// writer's clock seam, the model label, the elapsed epoch, the stderr-column
// probe, the idle-gap threshold, whether the indicator is enabled (the spinner
// gate), and whether the chrome colour is enabled (the round-057 grey `[Tool
// Output]` accent). It is the injected seam (a func-typed deps field) that keeps
// internal/cli free of internal/ui.
type ProgressFactory func(stream io.Writer, now func() time.Time, model string, epoch time.Time, columns func() int, idleGap time.Duration, indicatorEnabled, colour bool) TurnProgress

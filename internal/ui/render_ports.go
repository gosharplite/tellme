package ui

import (
	"io"
	"time"

	"github.com/gosharplite/tellme/internal/domain/history"
	"github.com/gosharplite/tellme/internal/domain/metrics"
	"github.com/gosharplite/tellme/internal/domain/render"
)

// Round 051 (R5.5 of #92; ADR 0020): the adapters that let internal/cli obtain
// its presentation through domain ports, so internal/cli names no internal/ui
// type. The BYTES (sanitize/fold/cap/format policy) stay owned HERE; the domain
// ports carry only the shapes.

// Lines is the production render.Lines implementation: a thin adapter over the
// pure status/tail formatters. Round 054 (ADR 0023): the adapter carries the
// chrome-colour flag the CLI resolved (the diagnostic stream is a terminal AND
// `-r` is off), so the pure formatters stay parameter-stable and the colour
// decision lives at the composition seam.
type Lines struct{ colour bool }

var _ render.Lines = Lines{}

// NewLines builds the production lines adapter with the round-054 colour flag.
func NewLines(colour bool) Lines { return Lines{colour: colour} }

// InputCaptured renders the input-capture acknowledgement.
func (Lines) InputCaptured(t time.Time) string { return FormatInputCaptured(t) }

// TurnOpening renders the frame's opening block.
func (Lines) TurnOpening(turn int, mode string) string { return FormatTurnOpening(turn, mode) }

// TurnGap renders the blank line separating the frame from the answer.
func (Lines) TurnGap() string { return FormatTurnGap() }

// PayloadMeasured renders the measured payload status line (round-054 green
// accents).
func (l Lines) PayloadMeasured(t time.Time, tokens, budget int, mode, model string) string {
	return formatPayloadMeasuredColour(t, tokens, budget, mode, model, l.colour)
}

// PayloadEstimate renders the estimated pre-flight payload line (round 057;
// ADR 0027): `+<delta> ~<tokens> tokens - <mode> - <model>` with the round-054
// green MODE accent.
func (l Lines) PayloadEstimate(t time.Time, tokens, delta int, mode, model string) string {
	return formatPayloadEstimateColour(t, tokens, delta, mode, model, l.colour)
}

// Metrics renders the per-turn metrics line.
func (Lines) Metrics(t time.Time, provider string, u metrics.UsageCounts) string {
	return FormatMetrics(t, provider, u)
}

// Ready renders the `╰─⠿ Ready` session summary (round-054 green session cost).
func (l Lines) Ready(lastCallCost, turnCost, sessionCost float64, sessionMiss, sessionHit, sessionOut int, hitRate float64) string {
	return formatReadyColour(lastCallCost, turnCost, sessionCost, sessionMiss, sessionHit, sessionOut, hitRate, l.colour)
}

// ToolReason renders the `[Tool Reason]` line (round-054 whole-line green).
func (l Lines) ToolReason(t time.Time, reason string) string {
	return formatToolReasonColour(t, reason, l.colour)
}

// ToolUsage renders the offline per-tool roll-up.
func (Lines) ToolUsage(rows []history.ToolUsageRow) string { return FormatToolUsage(rows) }

// DefaultToolOutputIdleGap is the `[Tool Output]` idle-gap default.
func (Lines) DefaultToolOutputIdleGap() time.Duration { return DefaultToolOutputIdleGap }

// Answer is the production render.Answer implementation: a thin adapter over the
// markdown Renderer.
type Answer struct{ r *Renderer }

var _ render.Answer = Answer{}

// NewAnswer builds the production answer renderer.
func NewAnswer() Answer { return Answer{r: NewRenderer()} }

// Render renders markdown to ANSI.
func (a Answer) Render(markdown string, width int) (string, bool) { return a.r.Render(markdown, width) }

// WarnDegraded emits the one-time degradation warning.
func (a Answer) WarnDegraded(w io.Writer) { a.r.WarnDegraded(w) }

// NewTurnProgress builds a turn's coupled presentation objects: the progress
// spinner (nil when indicatorEnabled is false) and the `[Tool Output]`
// coordinator (always present; it renders unconditionally). It is bound at the
// composition root as the deps.ProgressFactory seam.
func NewTurnProgress(spec render.ProgressSpec, m metrics.SystemMetricsProvider) render.TurnProgress {
	var sp *Spinner
	if spec.Spinner {
		sp = NewSpinner(spec.Stream, spec.Model, spec.Epoch, m, spec.Columns)
	}
	coord := NewToolOutputCoordinator(spec.Stream, spec.Now, sp, spec.IdleGap, spec.Colour)
	var ind render.Indicator
	if sp != nil {
		ind = sp
	}
	return render.TurnProgress{Indicator: ind, ToolOutput: coord}
}

// ToolLines returns the production agentport.ToolLineRenderer (the loop's
// four-line renderer) as a value the composition root injects, with the round-054
// chrome-colour flag.
func ToolLines(colour bool) ToolLineRenderer { return ToolLineRenderer{colour: colour} }

// Compile-time conformance: the spinner is the CLI's progress indicator port.
var _ render.Indicator = (*Spinner)(nil)

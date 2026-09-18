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
// pure status/tail formatters.
type Lines struct{}

var _ render.Lines = Lines{}

// InputCaptured renders the input-capture acknowledgement.
func (Lines) InputCaptured(t time.Time) string { return FormatInputCaptured(t) }

// TurnOpening renders the frame's opening block.
func (Lines) TurnOpening(turn int, mode string) string { return FormatTurnOpening(turn, mode) }

// TurnGap renders the blank line separating the frame from the answer.
func (Lines) TurnGap() string { return FormatTurnGap() }

// PayloadStatus renders one payload status line.
func (Lines) PayloadStatus(t time.Time, tokens, budget int, mode, model string, estimated bool) string {
	return FormatPayloadStatus(t, tokens, budget, mode, model, estimated)
}

// Metrics renders the per-turn metrics line.
func (Lines) Metrics(t time.Time, provider string, u metrics.UsageCounts) string {
	return FormatMetrics(t, provider, u)
}

// Ready renders the `╰─⠿ Ready` session summary.
func (Lines) Ready(lastCallCost, turnCost, sessionCost float64, sessionMiss, sessionHit, sessionOut int, hitRate float64) string {
	return FormatReady(lastCallCost, turnCost, sessionCost, sessionMiss, sessionHit, sessionOut, hitRate)
}

// ToolReason renders the `[Tool Reason]` line.
func (Lines) ToolReason(t time.Time, reason string) string { return FormatToolReason(t, reason) }

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
func NewTurnProgress(stream io.Writer, now func() time.Time, model string, epoch time.Time, m metrics.SystemMetricsProvider, columns func() int, idleGap time.Duration, indicatorEnabled bool) render.TurnProgress {
	var sp *Spinner
	if indicatorEnabled {
		sp = NewSpinner(stream, model, epoch, m, columns)
	}
	coord := NewToolOutputCoordinator(stream, now, sp, idleGap)
	var ind render.Indicator
	if sp != nil {
		ind = sp
	}
	return render.TurnProgress{Indicator: ind, ToolOutput: coord}
}

// ToolLines returns the production agentport.ToolLineRenderer (the loop's
// four-line renderer) as a value the composition root injects.
func ToolLines() ToolLineRenderer { return ToolLineRenderer{} }

// Compile-time conformance: the spinner is the CLI's progress indicator port.
var _ render.Indicator = (*Spinner)(nil)

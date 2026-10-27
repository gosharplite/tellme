package ui

import (
	"fmt"

	"github.com/gosharplite/tellme/internal/domain/render"
)

// FormatPayloadMeasured renders the per-turn MEASURED payload status line
// (round-009 research Decisions 3 & 4; round 057 TD-057-1 narrowed it to the
// measured shape only — the estimated/pre-flight shape is now separate):
//
//	[HH:MM:SS] Payload: <tokens>/<budget> tokens - <mode> - <model>
//
// Round 098 (ADR 0068): the figures arrive as a render.PayloadStatus, so the two
// adjacent ints and two adjacent strings cannot be swapped at a call site.
//
// The line is written to the diagnostic stream by the caller and deliberately
// carries NO `tellme: ` prefix, so the frozen class-phrase vocabulary is
// untouched (round-009 FR-014). The timestamp is supplied by the caller's
// injected clock seam so assertions stay deterministic.
func FormatPayloadMeasured(s render.PayloadStatus) string {
	return fmt.Sprintf("[%s] Payload: %d/%d tokens - %s - %s",
		formatClock(s.At), s.Tokens, s.Budget, s.Mode, s.Model)
}

// FormatPayloadEstimate renders the estimated pre-flight payload line (round
// 057; ADR 0027): it shows how much the estimate GREW over the previous estimate
// and drops the constant budget:
//
//	[HH:MM:SS] Payload: +<delta> ~<tokens> tokens - <mode> - <model>
//
// The delta is a signed difference (`+100`, `+0` when there is no predecessor, or
// `-1234` when the payload shrank). The measured line keeps the absolute
// `tokens/budget` form (FormatPayloadMeasured).
func FormatPayloadEstimate(s render.PayloadStatus) string {
	return fmt.Sprintf("[%s] Payload: %+d ~%d tokens - %s - %s",
		formatClock(s.At), s.Delta, s.Tokens, s.Mode, s.Model)
}

// formatPayloadEstimateColour is FormatPayloadEstimate with the round-054 green
// MODE accent (ADR 0023) — the estimated token number stays plain (the round-054
// measured-token accent does not apply here). The colour-off path returns
// FormatPayloadEstimate verbatim (the single plain entry point).
func formatPayloadEstimateColour(s render.PayloadStatus, colour bool) string {
	if !colour {
		return FormatPayloadEstimate(s)
	}
	return fmt.Sprintf("[%s] Payload: %+d ~%d tokens - %s - %s",
		formatClock(s.At), s.Delta, s.Tokens, green(s.Mode, true), s.Model)
}

// formatPayloadMeasuredColour is FormatPayloadMeasured with the round-054 green
// accents (ADR 0023): the MODE token is green and the measured token number is
// green. The colour-off path returns FormatPayloadMeasured verbatim (the single
// plain entry point), so the round-046 no-orphan rule holds.
func formatPayloadMeasuredColour(s render.PayloadStatus, colour bool) string {
	if !colour {
		return FormatPayloadMeasured(s)
	}
	tokenText := green(fmt.Sprintf("%d", s.Tokens), true)
	return fmt.Sprintf("[%s] Payload: %s/%d tokens - %s - %s",
		formatClock(s.At), tokenText, s.Budget, green(s.Mode, true), s.Model)
}

package ui

import (
	"fmt"
	"time"
)

// FormatPayloadMeasured renders the per-turn MEASURED payload status line
// (round-009 research Decisions 3 & 4; round 057 TD-057-1 narrowed it to the
// measured shape only — the estimated/pre-flight shape is now separate):
//
//	[HH:MM:SS] Payload: <tokens>/<budget> tokens - <mode> - <model>
//
// The line is written to the diagnostic stream by the caller and deliberately
// carries NO `tellme: ` prefix, so the frozen class-phrase vocabulary is
// untouched (round-009 FR-014). The timestamp is supplied by the caller's
// injected clock seam so assertions stay deterministic.
func FormatPayloadMeasured(t time.Time, tokens, budget int, mode, model string) string {
	return fmt.Sprintf("[%s] Payload: %d/%d tokens - %s - %s",
		formatClock(t), tokens, budget, mode, model)
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
func FormatPayloadEstimate(t time.Time, tokens, delta int, mode, model string) string {
	return fmt.Sprintf("[%s] Payload: %+d ~%d tokens - %s - %s",
		formatClock(t), delta, tokens, mode, model)
}

// formatPayloadEstimateColour is FormatPayloadEstimate with the round-054 green
// MODE accent (ADR 0023) — the estimated token number stays plain (the round-054
// measured-token accent does not apply here). The colour-off path returns
// FormatPayloadEstimate verbatim (the single plain entry point).
func formatPayloadEstimateColour(t time.Time, tokens, delta int, mode, model string, colour bool) string {
	if !colour {
		return FormatPayloadEstimate(t, tokens, delta, mode, model)
	}
	return fmt.Sprintf("[%s] Payload: %+d ~%d tokens - %s - %s",
		formatClock(t), delta, tokens, green(mode, true), model)
}

// formatPayloadMeasuredColour is FormatPayloadMeasured with the round-054 green
// accents (ADR 0023): the MODE token is green and the measured token number is
// green. The colour-off path returns FormatPayloadMeasured verbatim (the single
// plain entry point), so the round-046 no-orphan rule holds.
func formatPayloadMeasuredColour(t time.Time, tokens, budget int, mode, model string, colour bool) string {
	if !colour {
		return FormatPayloadMeasured(t, tokens, budget, mode, model)
	}
	tokenText := green(fmt.Sprintf("%d", tokens), true)
	return fmt.Sprintf("[%s] Payload: %s/%d tokens - %s - %s",
		formatClock(t), tokenText, budget, green(mode, true), model)
}

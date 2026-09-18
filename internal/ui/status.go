package ui

import (
	"fmt"
	"time"
)

// FormatPayloadStatus renders one per-turn payload status line (round-009
// research Decisions 3 & 4):
//
//	[HH:MM:SS] Payload: ~<tokens>/<budget> tokens - <mode> - <model>   (estimated, pre-flight)
//	[HH:MM:SS] Payload: <tokens>/<budget> tokens - <mode> - <model>    (measured, post-turn)
//
// `estimated` prefixes the count with `~`. The line is written to the diagnostic
// stream by the caller and deliberately carries NO `tellme: ` prefix, so the
// frozen class-phrase vocabulary is untouched (round-009 FR-014). The timestamp
// is supplied by the caller's injected clock seam so assertions stay
// deterministic.
func FormatPayloadStatus(t time.Time, tokens, budget int, mode, model string, estimated bool) string {
	return formatPayloadStatusColour(t, tokens, budget, mode, model, estimated, false)
}

// formatPayloadStatusColour is FormatPayloadStatus with the round-054 green
// accents (ADR 0023): the MODE token is green in BOTH lines, and the MEASURED
// token number is green (the `~`-estimated pre-flight number stays plain). The
// plain (enabled=false) path is byte-identical to the original format.
func formatPayloadStatusColour(t time.Time, tokens, budget int, mode, model string, estimated bool, colour bool) string {
	tilde := ""
	if estimated {
		tilde = "~"
	}
	tokenText := fmt.Sprintf("%d", tokens)
	if !estimated {
		tokenText = green(tokenText, colour)
	}
	return fmt.Sprintf("[%s] Payload: %s%s/%d tokens - %s - %s",
		formatClock(t), tilde, tokenText, budget, green(mode, colour), model)
}

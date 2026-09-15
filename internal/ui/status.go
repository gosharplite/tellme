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
	tilde := ""
	if estimated {
		tilde = "~"
	}
	return fmt.Sprintf("[%s] Payload: %s%d/%d tokens - %s - %s",
		formatClock(t), tilde, tokens, budget, mode, model)
}

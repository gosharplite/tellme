package ui

import (
	"fmt"
	"time"

	"github.com/gosharplite/tellme/internal/domain/metrics"
)

// FormatMetrics renders the round-018 per-turn metrics line:
//
//	[HH:MM:SS] [<provider>] M: <miss> H: <cached> C: <completion> Th: <thinking>
//
// The `Th:` segment is ALWAYS rendered, including `Th: 0` (operator decision — a
// deviation from the reference's suppress-when-0). The `[<provider>]` bracket is
// rendered unconditionally (format stability). The timestamp comes from the
// caller's injected clock seam.
func FormatMetrics(t time.Time, provider string, u metrics.UsageCounts) string {
	return fmt.Sprintf("[%s] [%s] M: %d H: %d C: %d Th: %d",
		formatClock(t), provider, u.Miss, u.Hit, u.Completion, u.Thinking)
}

// FormatReady renders the round-018 `╰─⠿ Ready` session summary:
//
//	╰─⠿ Ready ($<lastCall> $<turn> $<session> - M: <sMiss> H: <sHit> O: <sOut> - <hit%>%)
//
// The three costs are formatted `$%.4f` (last request / whole turn / session);
// M/H/O are the SESSION-cumulative miss/cached/output token totals; the hit-rate
// is `%.1f%%` (research Decisions 3 & 7). The line carries no timestamp.
func FormatReady(lastCallCost, turnCost, sessionCost float64, sessionMiss, sessionHit, sessionOut int, hitRate float64) string {
	return fmt.Sprintf("╰─⠿ Ready ($%.4f $%.4f $%.4f - M: %d H: %d O: %d - %.1f%%)",
		lastCallCost, turnCost, sessionCost, sessionMiss, sessionHit, sessionOut, hitRate)
}

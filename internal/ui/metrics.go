package ui

import (
	"fmt"
	"time"
)

// UsageCounts is the per-call token breakdown shown on the metrics line
// (round-018): the derived miss (M) and the reported cached (H), completion (C),
// and reasoning (Th) tokens.
type UsageCounts struct {
	Miss       int
	Hit        int
	Completion int
	Thinking   int
}

// FormatMetrics renders the round-018 per-turn metrics line:
//
//	[HH:MM:SS] [<provider>] M: <miss> H: <cached> C: <completion> Th: <thinking>
//
// The `Th:` segment is ALWAYS rendered, including `Th: 0` (operator decision — a
// deviation from the reference's suppress-when-0). The `[<provider>]` bracket is
// rendered unconditionally (format stability). The timestamp comes from the
// caller's injected clock seam.
func FormatMetrics(t time.Time, provider string, u UsageCounts) string {
	return fmt.Sprintf("[%s] [%s] M: %d H: %d C: %d Th: %d",
		t.Format("15:04:05"), provider, u.Miss, u.Hit, u.Completion, u.Thinking)
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

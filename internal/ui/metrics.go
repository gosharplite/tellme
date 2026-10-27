package ui

import (
	"fmt"
	"time"

	"github.com/gosharplite/tellme/internal/domain/metrics"
	"github.com/gosharplite/tellme/internal/domain/render"
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
//
// Round 098 (ADR 0068): the seven positional figures are one render.ReadyStats —
// three adjacent float64s and three adjacent ints could otherwise be swapped at a
// call site and still compile.
func FormatReady(stats render.ReadyStats) string {
	return fmt.Sprintf("╰─⠿ Ready ($%.4f $%.4f $%.4f - M: %d H: %d O: %d - %.1f%%)",
		stats.LastCallCost, stats.TurnCost, stats.SessionCost, stats.SessionMiss, stats.SessionHit, stats.SessionOut, stats.HitRate)
}

// formatReadyColour is FormatReady with the round-054 green accent (ADR 0023):
// only the THIRD (session) cost is green (reference parity — tell-me-go greens
// the session cost). The colour-off path returns FormatReady verbatim.
func formatReadyColour(stats render.ReadyStats, colour bool) string {
	if !colour {
		return FormatReady(stats)
	}
	return fmt.Sprintf("╰─⠿ Ready ($%.4f $%.4f %s - M: %d H: %d O: %d - %.1f%%)",
		stats.LastCallCost, stats.TurnCost, green(fmt.Sprintf("$%.4f", stats.SessionCost), true),
		stats.SessionMiss, stats.SessionHit, stats.SessionOut, stats.HitRate)
}

package ui

import (
	"fmt"
	"strings"
	"time"
)

// FormatToolLog renders the round-022 single-line tool-loop diagnostic written to
// the diagnostic stream (`stderr`) for one tool call:
//
//	[HH:MM:SS] [Tool] <tool name> - <reason>   (the call states a reason)
//	[HH:MM:SS] [Tool] <tool name>              (no reason — no dangling separator)
//
// The raw call arguments and its result are deliberately NOT included (the
// operator-chosen shape, round-022 research Decision 1/3). The reason is folded
// to a single line — newlines collapsed, ends trimmed — so a multi-line or
// whitespace-only reason cannot break the one-line-per-call contract (round-022
// FR-005): an effectively-empty reason takes the no-tail branch. The timestamp
// comes from the caller's injected clock seam, via the shared formatClock token.
func FormatToolLog(t time.Time, name, reason string) string {
	reason = strings.TrimSpace(oneLine(reason))
	if reason == "" {
		return fmt.Sprintf("[%s] [Tool] %s", formatClock(t), name)
	}
	return fmt.Sprintf("[%s] [Tool] %s - %s", formatClock(t), name, reason)
}

// oneLine folds newlines so a multi-line reason cannot break the single-line log
// contract (round-022 FR-005 — the guarantee lived in the pre-022 `reasonSegment`
// and is re-pinned here in the pure formatter).
func oneLine(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\n", " "), "\r", " ")
}

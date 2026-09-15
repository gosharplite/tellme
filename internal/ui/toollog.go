package ui

import (
	"fmt"
	"time"
)

// FormatToolLog renders the round-022 single-line tool-loop diagnostic written to
// the diagnostic stream (`stderr`) for one tool call:
//
//	[HH:MM:SS] [Tool] <tool name> - <reason>   (the call states a reason)
//	[HH:MM:SS] [Tool] <tool name>              (no reason — no dangling separator)
//
// The raw call arguments and its result are deliberately NOT included (the
// operator-chosen shape, round-022 research Decision 1/3); the ` - <reason>` tail
// is omitted entirely when reason is empty (Decision 4). The timestamp comes from
// the caller's injected clock seam, rendered through the shared formatClock token.
func FormatToolLog(t time.Time, name, reason string) string {
	if reason == "" {
		return fmt.Sprintf("[%s] [Tool] %s", formatClock(t), name)
	}
	return fmt.Sprintf("[%s] [Tool] %s - %s", formatClock(t), name, reason)
}

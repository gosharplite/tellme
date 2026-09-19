package ui

import (
	"fmt"
	"strings"

	"github.com/gosharplite/tellme/internal/domain/history"
)

// FormatToolUsage renders the offline per-tool roll-up (round 026): a header plus
// one line per tool in the supplied order (the report's recordable order — the
// union of the base and capability-gated sets, round 062), each carrying its
// lifetime invocation tally across every session.
//
//	tool usage (all sessions):
//	list_files: total=1 ok=1 error=0 timeout=0
//	read_files: total=3 ok=2 error=1 timeout=0
//
// The format is a pure function (no I/O, no clock) so it is unit-pinnable
// (round 026 T017), and it is the single source the E2E Report Thens parse.
func FormatToolUsage(rows []history.ToolUsageRow) string {
	var b strings.Builder
	b.WriteString("tool usage (all sessions):\n")
	for _, r := range rows {
		fmt.Fprintf(&b, "%s: total=%d ok=%d error=%d timeout=%d\n", r.Tool, r.Total(), r.OK(), r.Error(), r.Timeout())
	}
	return b.String()
}

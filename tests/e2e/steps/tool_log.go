package steps

import (
	"strings"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// Round-022 tool-log helpers. The tool-loop diagnostic line is
// `[HH:MM:SS] [Tool] <name> - <reason>` (the ` - <reason>` tail is omitted
// entirely when the call states no reason). Kept in ONE file so the per-sentence
// step files stay independent (Zero Shared Edits), mirroring wire_tools.go /
// ordering_witness.go / payload_status.go.

// toolLog is one parsed tool-loop log line.
type toolLog struct {
	name      string
	reason    string
	hasReason bool
}

// parseToolLog parses a line of the form `[HH:MM:SS] [Tool] <name>[ - <reason>]`.
// It returns ok=false for any other line.
func parseToolLog(line string) (toolLog, bool) {
	const marker = "] [Tool] "
	i := strings.Index(line, marker)
	if i < 0 {
		return toolLog{}, false
	}
	rest := strings.TrimRight(line[i+len(marker):], "\r")
	if k := strings.Index(rest, " - "); k >= 0 {
		return toolLog{name: rest[:k], reason: rest[k+3:], hasReason: true}, true
	}
	return toolLog{name: rest}, true
}

// toolLogs returns every tool-loop log line in the captured stderr, in order.
func toolLogs(stderr string) []toolLog {
	out := make([]toolLog, 0)
	for _, line := range strings.Split(stderr, "\n") {
		if tl, ok := parseToolLog(line); ok {
			out = append(out, tl)
		}
	}
	return out
}

// executedToolCount returns the number of tool calls the run executed, derived
// from the run's final provider request (which carries one role:"tool" message
// per executed tool). It is the expected number of tool-loop log lines.
func executedToolCount(f *fakeprovider.Provider) int {
	if f == nil {
		return 0
	}
	n := 0
	for _, m := range f.MessagesAt(-1) {
		if role, _ := m["role"].(string); role == "tool" {
			n++
		}
	}
	return n
}

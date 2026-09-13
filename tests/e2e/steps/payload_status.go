package steps

import (
	"regexp"
	"strconv"
	"strings"
)

// Round-009 payload-status helpers: parse the per-turn payload status lines the
// CLI writes to the diagnostic stream (stderr). Kept in ONE file so the
// per-sentence step files stay independent (Zero Shared Edits), mirroring
// wire_tools.go.
//
// The line shape is pinned by specs/truth/features/cli/chat/dsl.md:
//
//	[HH:MM:SS] Payload: ~<est>/<budget> tokens - <mode> - <model>   (pre-flight)
//	[HH:MM:SS] Payload: <actual>/<budget> tokens - <mode> - <model> (post-turn)
var (
	reEstimatedPayload = regexp.MustCompile(`\[[0-9]{2}:[0-9]{2}:[0-9]{2}\] Payload: ~([0-9]+)/([0-9]+) tokens - (\S+) - (\S+)`)
	reMeasuredPayload  = regexp.MustCompile(`\[[0-9]{2}:[0-9]{2}:[0-9]{2}\] Payload: ([0-9]+)/([0-9]+) tokens - (\S+) - (\S+)`)
)

// hasEstimatedPayloadStatus reports whether stderr carries a pre-flight
// estimated (`~`) payload status line.
func hasEstimatedPayloadStatus(stderr string) bool { return reEstimatedPayload.MatchString(stderr) }

// hasMeasuredPayloadStatus reports whether stderr carries a post-turn measured
// payload status line (no `~`).
func hasMeasuredPayloadStatus(stderr string) bool { return reMeasuredPayload.MatchString(stderr) }

// payloadStatusBudget returns the `<max>` budget of the first payload status
// line (estimated or measured) in stderr, and whether one was found.
func payloadStatusBudget(stderr string) (int, bool) {
	if m := reEstimatedPayload.FindStringSubmatch(stderr); m != nil {
		n, _ := strconv.Atoi(m[2])
		return n, true
	}
	if m := reMeasuredPayload.FindStringSubmatch(stderr); m != nil {
		n, _ := strconv.Atoi(m[2])
		return n, true
	}
	return 0, false
}

// hasPayloadStatusLine reports whether stderr carries any payload status line.
func hasPayloadStatusLine(stderr string) bool {
	return strings.Contains(stderr, "Payload: ")
}

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
//	[HH:MM:SS] Payload: +<delta> ~<est> tokens - <mode> - <model>   (pre-flight; round 057)
//	[HH:MM:SS] Payload: <actual>/<budget> tokens - <mode> - <model> (post-turn)
//
// Round 057 (ADR 0027): the estimated line shows the increment over the previous
// estimate (a signed `+`/`-`) and drops the `/budget`; the measured line is
// unchanged. Groups: estimated = (delta, tokens, mode, model).
var (
	reEstimatedPayload = regexp.MustCompile(`\[[0-9]{2}:[0-9]{2}:[0-9]{2}\] Payload: ([+-][0-9]+) ~([0-9]+) tokens - (\S+) - (\S+)`)
	reMeasuredPayload  = regexp.MustCompile(`\[[0-9]{2}:[0-9]{2}:[0-9]{2}\] Payload: ([0-9]+)/([0-9]+) tokens - (\S+) - (\S+)`)
)

// hasEstimatedPayloadStatus reports whether stderr carries a pre-flight
// estimated (`~`) payload status line.
func hasEstimatedPayloadStatus(stderr string) bool { return reEstimatedPayload.MatchString(stderr) }

// hasMeasuredPayloadStatus reports whether stderr carries a post-turn measured
// payload status line (no `~`).
func hasMeasuredPayloadStatus(stderr string) bool { return reMeasuredPayload.MatchString(stderr) }

// payloadStatusBudget returns the `<max>` budget of the first MEASURED payload
// status line in stderr, and whether one was found. Round 057 (ADR 0027): the
// estimated line no longer carries a budget, so only the measured line does.
func payloadStatusBudget(stderr string) (int, bool) {
	if m := reMeasuredPayload.FindStringSubmatch(stderr); m != nil {
		n, _ := strconv.Atoi(m[2])
		return n, true
	}
	return 0, false
}

// estimatedPayloadValue returns the `<est>` of the first pre-flight estimated
// payload status line in stderr, and whether one was found (round-011 estimate
// assertions). Round 057 (ADR 0027): the estimate is group 2 (group 1 is the
// signed delta).
func estimatedPayloadValue(stderr string) (int, bool) {
	if m := reEstimatedPayload.FindStringSubmatch(stderr); m != nil {
		n, _ := strconv.Atoi(m[2])
		return n, true
	}
	return 0, false
}

// estimatedPayloadDelta returns the signed delta of the first pre-flight estimate
// line in stderr, and whether one was found (round-057 increment assertions).
func estimatedPayloadDelta(stderr string) (int, bool) {
	if m := reEstimatedPayload.FindStringSubmatch(stderr); m != nil {
		n, err := strconv.Atoi(m[1])
		return n, err == nil
	}
	return 0, false
}

// hasPayloadStatusLine reports whether stderr carries any payload status line.
func hasPayloadStatusLine(stderr string) bool {
	return strings.Contains(stderr, "Payload: ")
}

// isEstimatePayloadLine reports whether a diagnostic line is a pre-flight
// payload ESTIMATE line. Round 057 (ADR 0027): the estimate is
// `Payload: +<delta> ~<tokens> …` (a signed increment before the `~`); the
// measured line is `Payload: <tokens>/<budget> …`.
func isEstimatePayloadLine(line string) bool {
	return strings.Contains(line, "Payload: +") || strings.Contains(line, "Payload: -")
}

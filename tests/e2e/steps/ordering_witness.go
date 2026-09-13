package steps

import (
	"regexp"

	"github.com/gosharplite/tellme/tests/e2e/harness"
)

// Round-010 cross-stream ordering helpers. The ordering assertions run against
// the MERGED capture (sc.merged) — the single ordered buffer that records the
// interleave of stdout and stderr, the witness round 009 lacked. Kept in ONE
// file so the per-sentence step files stay independent (Zero Shared Edits),
// mirroring payload_status.go / wire_tools.go.

// mergedView returns the scenario's merged (stdout+stderr) capture with ANSI
// escape sequences stripped. The renderer interleaves SGR codes between words,
// so the answer text is only searchable contiguously after stripping; the
// ordering (byte positions) is preserved by stripping.
func mergedView(sc *scenarioContext) string { return harness.StripANSI(sc.merged) }

// firstMatchIndex returns the byte index of the first match of re in s, or -1.
func firstMatchIndex(re *regexp.Regexp, s string) int {
	if loc := re.FindStringIndex(s); loc != nil {
		return loc[0]
	}
	return -1
}

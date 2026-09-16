package steps

import (
	"regexp"
	"strconv"
	"strings"
)

// Round-034 decomposed tool-call log helpers (ADR 0005). Parse the diagnostic
// stream's decomposed rendering:
//
//	[HH:MM:SS] [Tool Engine] Step <i>/<m>
//	[HH:MM:SS] [Tool Reason] <reason>
//	[HH:MM:SS] [Tool Action] <tool>(<sorted k: v>)
//	[HH:MM:SS] [Tool Output] Executing... (Output shown below)
//	[HH:MM:SS] [Tool Result] <tool>: <snippet>
//	╭─⠿ Turn <N> - <mode>   (a per-AI-endpoint-call frame)
//	╰─⠿ Ready (…)           (a per-call tail)
//
// Kept in ONE file so the per-sentence step files (step_r034_t0NN_*) stay
// independent (Zero Shared Edits), mirroring tool_log.go / wire_tools.go.

const (
	toolEngineMarker = "] [Tool Engine] "
	toolReasonMarker = "] [Tool Reason] "
	toolActionMarker = "] [Tool Action] "
	toolResultMarker = "] [Tool Result] "
	toolOutputMarker = "] [Tool Output] "
	turnHeaderMarker = "╭─⠿ Turn "
	readyMarker      = "╰─⠿ Ready "
)

// stderrLines splits the captured diagnostic stream into lines.
func stderrLines(stderr string) []string { return strings.Split(stderr, "\n") }

// indexOfLines returns the line indexes carrying substr.
func indexOfLines(stderr, substr string) []int {
	var idx []int
	for i, l := range stderrLines(stderr) {
		if strings.Contains(l, substr) {
			idx = append(idx, i)
		}
	}
	return idx
}

// hasToolEngineLine reports whether stderr carries at least one decomposed
// `[Tool Engine] Step i/M` line (T005).
func hasToolEngineLine(stderr string) bool { return strings.Contains(stderr, toolEngineMarker) }

// toolEngineLineCount counts the `[Tool Engine]` lines (T018).
func toolEngineLineCount(stderr string) int { return len(indexOfLines(stderr, toolEngineMarker)) }

// lastToolEngineIndex returns the line index of the last `[Tool Engine]` line,
// or -1 (T019).
func lastToolEngineIndex(stderr string) int {
	idx := indexOfLines(stderr, toolEngineMarker)
	if len(idx) == 0 {
		return -1
	}
	return idx[len(idx)-1]
}

// toolActionArgsFor returns, for every `[Tool Action] <tool>(…)` line naming
// tool, the argument body between the parentheses (T006/T007/T009).
func toolActionArgsFor(stderr, tool string) []string {
	var out []string
	prefix := tool + "("
	for _, l := range stderrLines(stderr) {
		i := strings.Index(l, toolActionMarker)
		if i < 0 {
			continue
		}
		rest := strings.TrimRight(l[i+len(toolActionMarker):], "\r")
		if !strings.HasPrefix(rest, prefix) {
			continue
		}
		body := strings.TrimSuffix(rest[len(prefix):], ")")
		out = append(out, body)
	}
	return out
}

// firstToolActionIndex returns the line index of the first `[Tool Action] <tool>(…)`
// line naming tool, or -1 (T009).
func firstToolActionIndex(stderr, tool string) int {
	prefix := tool + "("
	for i, l := range stderrLines(stderr) {
		j := strings.Index(l, toolActionMarker)
		if j < 0 {
			continue
		}
		if strings.HasPrefix(strings.TrimRight(l[j+len(toolActionMarker):], "\r"), prefix) {
			return i
		}
	}
	return -1
}

// hasToolResultFor reports whether stderr carries a `[Tool Result] <tool>: …`
// line for tool (T007).
func hasToolResultFor(stderr, tool string) bool {
	return strings.Contains(stderr, toolResultMarker+tool+": ")
}

// hasToolOutputBlock reports whether stderr carries the `[Tool Output]` header
// (T010/T011/T016).
func hasToolOutputBlock(stderr string) bool { return strings.Contains(stderr, toolOutputMarker) }

// toolOutputBlockIndexes returns the [headerIdx, lastIdx] line range of the
// first `[Tool Output]` block (the header's line through the last `[Tool Output]`
// line), or (-1,-1) when absent (T016).
func toolOutputBlockIndexes(stderr string) (int, int) {
	idx := indexOfLines(stderr, toolOutputMarker)
	if len(idx) == 0 {
		return -1, -1
	}
	return idx[0], idx[len(idx)-1]
}

// turnFrameNumbers returns the `<N>` of every `╭─⠿ Turn <N>` header, in order
// (T012/T013/T019).
func turnFrameNumbers(stderr string) []int {
	var out []int
	for _, l := range stderrLines(stderr) {
		i := strings.Index(l, turnHeaderMarker)
		if i < 0 {
			continue
		}
		rest := l[i+len(turnHeaderMarker):]
		j := strings.IndexByte(rest, ' ')
		if j < 0 {
			j = len(rest)
		}
		if n, err := strconv.Atoi(strings.TrimSpace(rest[:j])); err == nil {
			out = append(out, n)
		}
	}
	return out
}

// readyLineCount counts the `╰─⠿ Ready` tail lines (T014).
func readyLineCount(stderr string) int { return len(indexOfLines(stderr, readyMarker)) }

// estimatedPayloadLineCount counts the pre-flight `Payload: ~` lines (T017).
func estimatedPayloadLineCount(stderr string) int {
	return len(indexOfLines(stderr, "Payload: ~"))
}

// modelRequestCount returns the number of model requests the scenario's fake
// recorded (T012/T014/T017). 0 when no fake was started.
func modelRequestCount(sc *scenarioContext) int {
	f := sc.onlyFake()
	if f == nil {
		return 0
	}
	return f.RequestCount()
}

// spinnerStatusRe matches a live spinner status line (`<braille> Thinking…` /
// `<braille> Executing…`), used by the "spinner paused during a stream" pin
// (T016).
var spinnerStatusRe = regexp.MustCompile(`[\x{2800}-\x{28FF}] (Thinking|Executing)`)

// hasSpinnerStatusBetween reports whether any line in (after, before) exclusive
// carries a live spinner status line (T016).
func hasSpinnerStatusBetween(lines []string, after, before int) bool {
	for i := after + 1; i < before && i < len(lines); i++ {
		if spinnerStatusRe.MatchString(lines[i]) {
			return true
		}
	}
	return false
}

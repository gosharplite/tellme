package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T004 [BDD-ALIGN] — Then: the run reported the reason "{reason}" for the tool call "{tool}"
//
// Round 034 retargets the round-021/022 reason row from the retired single-line
// `[HH:MM:SS] [Tool] <name> - <reason>` shape to the decomposed round-034 shape:
// a `[HH:MM:SS] [Tool Reason] <reason>` line emitted for the call to `{tool}` —
// i.e. immediately before that call's `[Tool Action] <tool>(…)` line (FR-002).
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the run reported the reason "([^"]*)" for the tool call "([^"]*)"$`, thenReasonEcho)
	})
}

// thenReasonEcho (必查 呈現結果): the captured stderr carries a
// `[HH:MM:SS] [Tool Reason] {reason}` line whose next non-blank line is that
// call's `[HH:MM:SS] [Tool Action] {tool}(…)` line — the reason's own decomposed
// line, with no ` - ` tail, emitted for the call to {tool} (round 034 FR-002).
func thenReasonEcho(ctx context.Context, reason, tool string) error {
	sc := scenarioFrom(ctx)
	lines := strings.Split(sc.stderr, "\n")
	for i, line := range lines {
		if !reasonRowMatches(line, reason) {
			continue
		}
		for j := i + 1; j < len(lines); j++ {
			if strings.TrimSpace(lines[j]) == "" {
				continue
			}
			if actionRowMatches(lines[j], tool) {
				return nil
			}
			break
		}
	}
	return fmt.Errorf("standard error carried no `[Tool Reason] %s` line before the action for %q; stderr=%q", reason, tool, sc.stderr)
}

// reasonRowMatches reports whether line is a decomposed `[Tool Reason] <reason>`
// line carrying exactly reason.
func reasonRowMatches(line, reason string) bool {
	const marker = "] [Tool Reason] "
	i := strings.Index(line, marker)
	if i < 0 {
		return false
	}
	return strings.TrimRight(line[i+len(marker):], "\r") == reason
}

// actionRowMatches reports whether line is a decomposed
// `[Tool Action] <tool>(…)` line for the given tool.
func actionRowMatches(line, tool string) bool {
	const marker = "] [Tool Action] "
	i := strings.Index(line, marker)
	if i < 0 {
		return false
	}
	return strings.HasPrefix(strings.TrimRight(line[i+len(marker):], "\r"), tool+"(")
}

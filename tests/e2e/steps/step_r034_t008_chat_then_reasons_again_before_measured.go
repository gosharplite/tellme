package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T008 [BDD-RED] — Then: the run reported the reasons again before the measured payload status
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the run reported the reasons again before the measured payload status$`, thenReasonsAgainBeforeMeasured)
	})
}

// thenReasonsAgainBeforeMeasured (必查 呈現結果): stderr carries a grouped
// `[HH:MM:SS] [Tool Reason] …` line AFTER the last `[Tool Result]` line and
// BEFORE the measured (no `~`) payload status line — the per-call tail's
// re-emitted reasons (round 034 FR-005).
func thenReasonsAgainBeforeMeasured(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	lines := stderrLines(sc.stderr)

	lastResult := -1
	for i, l := range lines {
		if strings.Contains(l, toolResultMarker) {
			lastResult = i
		}
	}
	if lastResult < 0 {
		return fmt.Errorf("standard error carried no `[Tool Result]` line to anchor the grouped reasons; stderr=%q", sc.stderr)
	}
	measured := -1
	for i, l := range lines {
		if strings.Contains(l, "Payload: ") && !isEstimatePayloadLine(l) {
			measured = i
			break
		}
	}
	for i, l := range lines {
		if i <= lastResult {
			continue
		}
		if measured >= 0 && i >= measured {
			break
		}
		if strings.Contains(l, toolReasonMarker) {
			return nil
		}
	}
	return fmt.Errorf("standard error carried no grouped `[Tool Reason]` line before the measured payload status; stderr=%q", sc.stderr)
}

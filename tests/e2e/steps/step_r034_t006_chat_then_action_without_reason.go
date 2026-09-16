package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T006 [BDD-RED] — Then: the run reported the action for the tool call "{tool}" without the reason
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the run reported the action for the tool call "([^"]*)" without the reason$`, thenActionWithoutReason)
	})
}

// thenActionWithoutReason (必查 呈現結果): stderr carries a decomposed
// `[HH:MM:SS] [Tool Action] <tool>(…)` line whose argument list excludes the
// top-level `reason` key (round 034 FR-003).
func thenActionWithoutReason(ctx context.Context, tool string) error {
	sc := scenarioFrom(ctx)
	bodies := toolActionArgsFor(sc.stderr, tool)
	if len(bodies) == 0 {
		return fmt.Errorf("standard error carried no `[Tool Action] %s(…)` line; stderr=%q", tool, sc.stderr)
	}
	for _, body := range bodies {
		if !strings.Contains(body, "reason") {
			return nil
		}
	}
	return fmt.Errorf("every `[Tool Action] %s(…)` line carried the reason; stderr=%q", tool, sc.stderr)
}

package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T009 [BDD-RED] — Then: the run reported the action for the tool call "{tool_a}" before the action for the tool call "{tool_b}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the run reported the action for the tool call "([^"]*)" before the action for the tool call "([^"]*)"$`, thenActionBeforeAction)
	})
}

// thenActionBeforeAction (必查 呈現結果): stderr carries a decomposed
// `[Tool Action] {tool_a}(…)` line BEFORE a `[Tool Action] {tool_b}(…)` line —
// the calls are rendered in the order they were made (round 034 US1).
func thenActionBeforeAction(ctx context.Context, toolA, toolB string) error {
	sc := scenarioFrom(ctx)
	ia := firstToolActionIndex(sc.stderr, toolA)
	ib := firstToolActionIndex(sc.stderr, toolB)
	if ia < 0 {
		return fmt.Errorf("no `[Tool Action] %s(…)` line; stderr=%q", toolA, sc.stderr)
	}
	if ib < 0 {
		return fmt.Errorf("no `[Tool Action] %s(…)` line; stderr=%q", toolB, sc.stderr)
	}
	if ia >= ib {
		return fmt.Errorf("the action for %q (line %d) was not reported before %q (line %d); stderr=%q", toolA, ia, toolB, ib, sc.stderr)
	}
	return nil
}

package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T007 [BDD-RED] — Then: the run reported the result for the tool call "{tool}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the run reported the result for the tool call "([^"]*)"$`, thenResultForCall)
	})
}

// thenResultForCall (必查 呈現結果): stderr carries a decomposed
// `[HH:MM:SS] [Tool Result] <tool>: <snippet>` line (round 034 FR-004).
func thenResultForCall(ctx context.Context, tool string) error {
	sc := scenarioFrom(ctx)
	if !hasToolResultFor(sc.stderr, tool) {
		return fmt.Errorf("standard error carried no `[Tool Result] %s: …` line; stderr=%q", tool, sc.stderr)
	}
	return nil
}

package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T010 [BDD-RED] — Then: the run streamed the command's output on its diagnostic output
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the run streamed the command's output on its diagnostic output$`, thenStreamedOutput)
	})
}

// thenStreamedOutput (必查 呈現結果): stderr carries the live `[Tool Output]`
// block — the header line plus at least one streamed output line (round 034
// FR-010).
func thenStreamedOutput(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if !hasToolOutputBlock(sc.stderr) {
		return fmt.Errorf("standard error carried no `[Tool Output]` block; stderr=%q", sc.stderr)
	}
	if len(indexOfLines(sc.stderr, toolOutputMarker)) < 2 {
		return fmt.Errorf("the `[Tool Output]` block carried no streamed line beyond the header; stderr=%q", sc.stderr)
	}
	return nil
}

package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T005 [BDD-RED] — Then: the run reported the tool step marker
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the run reported the tool step marker$`, thenToolStepMarker)
	})
}

// thenToolStepMarker (必查 呈現結果): stderr carries a decomposed
// `[HH:MM:SS] [Tool Engine] Step <i>/<M>` line (round 034 FR-001).
func thenToolStepMarker(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if !hasToolEngineLine(sc.stderr) {
		return fmt.Errorf("standard error carried no `[Tool Engine] Step i/M` line; stderr=%q", sc.stderr)
	}
	return nil
}

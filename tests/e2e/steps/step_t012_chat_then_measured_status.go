package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T012 — Then: tellme reports the measured payload status for the turn
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme reports the measured payload status for the turn$`, thenMeasuredPayloadStatus)
	})
}

// thenMeasuredPayloadStatus (必查 呈現結果): the captured standard error carries a
// post-turn measured payload status line (no `~`).
func thenMeasuredPayloadStatus(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if !hasMeasuredPayloadStatus(sc.stderr) {
		return fmt.Errorf("standard error carried no measured payload status line; stderr=%q", sc.stderr)
	}
	return nil
}

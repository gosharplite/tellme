package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T013 — Then: tellme reports no measured payload status
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme reports no measured payload status$`, thenNoMeasuredPayloadStatus)
	})
}

// thenNoMeasuredPayloadStatus (必查 呈現結果): the captured standard error carries
// no post-turn measured payload status line.
func thenNoMeasuredPayloadStatus(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if hasMeasuredPayloadStatus(sc.stderr) {
		return fmt.Errorf("standard error carried a measured payload status line; stderr=%q", sc.stderr)
	}
	return nil
}

package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T011 — Then: tellme reports the estimated payload status for the turn
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme reports the estimated payload status for the turn$`, thenEstimatedPayloadStatus)
	})
}

// thenEstimatedPayloadStatus (必查 呈現結果): the captured standard error carries
// a pre-flight estimated (`~`) payload status line, and it is NOT written to
// standard output.
func thenEstimatedPayloadStatus(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if !hasEstimatedPayloadStatus(sc.stderr) {
		return fmt.Errorf("standard error carried no estimated payload status line; stderr=%q", sc.stderr)
	}
	if strings.Contains(sc.stdout, "Payload:") {
		return fmt.Errorf("the payload status leaked onto standard output: %q", sc.stdout)
	}
	return nil
}

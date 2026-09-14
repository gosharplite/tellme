package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T003 — Then: the input capture is announced for the turn
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the input capture is announced for the turn$`, thenInputCaptureAnnounced)
	})
}

// thenInputCaptureAnnounced (必查 呈現結果): the diagnostic stream carries the
// input-capture acknowledgement `[HH:MM:SS] Input captured. Processing...`, and
// it is NOT written to standard output.
func thenInputCaptureAnnounced(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if !hasTurnAck(sc.stderr) {
		return fmt.Errorf("the diagnostic stream carried no input-capture acknowledgement; stderr=%q", sc.stderr)
	}
	if hasTurnAck(sc.stdout) {
		return fmt.Errorf("the input-capture acknowledgement leaked to standard output; stdout=%q", sc.stdout)
	}
	return nil
}

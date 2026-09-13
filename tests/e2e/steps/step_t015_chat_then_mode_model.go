package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T015 — Then: the payload status names the active mode and model
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the payload status names the active mode and model$`, thenPayloadNamesModeModel)
	})
}

// thenPayloadNamesModeModel (必查 呈現結果): a reported payload status line ends
// with ` - <mode> - <model>` where both are non-empty (the regexes require
// non-space tokens), naming the run's effective mode and the provider's
// configured MODEL attribute (TD-2).
func thenPayloadNamesModeModel(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if !hasEstimatedPayloadStatus(sc.stderr) && !hasMeasuredPayloadStatus(sc.stderr) {
		return fmt.Errorf("standard error carried no payload status line naming the mode and model; stderr=%q", sc.stderr)
	}
	return nil
}

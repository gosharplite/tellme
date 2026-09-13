package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T014 — Then: the estimated payload matches the previous run's
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the estimated payload matches the previous run's$`, thenEstimateMatchesPrevious)
	})
}

// thenEstimateMatchesPrevious (必查 呈現結果): the current pre-flight estimate
// `~<n>` equals the stored previous estimate.
func thenEstimateMatchesPrevious(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if !sc.previousEstimateSet {
		return fmt.Errorf("no previous estimate was recorded (Given: a previous run …)")
	}
	n, ok := estimatedPayloadValue(sc.stderr)
	if !ok {
		return fmt.Errorf("standard error carried no estimated payload status line; stderr=%q", sc.stderr)
	}
	if n != sc.previousEstimate {
		return fmt.Errorf("the estimated payload %d must equal the previous run's %d", n, sc.previousEstimate)
	}
	return nil
}

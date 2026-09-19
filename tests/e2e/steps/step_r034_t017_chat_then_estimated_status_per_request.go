package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T017 [BDD-RED] — Then: each model request reports an estimated payload status
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^each model request reports an estimated payload status$`, thenEstimatedPerRequest)
	})
}

// thenEstimatedPerRequest (必查 呈現結果): the count of pre-flight `Payload: +`
// lines equals the number of model requests — every call reports its own
// estimate (round 034 FR-010a).
func thenEstimatedPerRequest(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	want := modelRequestCount(sc)
	if want == 0 {
		return fmt.Errorf("the scenario recorded no model request (no fake request); stderr=%q", sc.stderr)
	}
	got := estimatedPayloadLineCount(sc.stderr)
	if got != want {
		return fmt.Errorf("expected one estimated payload status per model request: got %d for %d requests; stderr=%q", got, want, sc.stderr)
	}
	return nil
}

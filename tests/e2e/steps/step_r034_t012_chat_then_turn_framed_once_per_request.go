package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T012 [BDD-RED] — Then: the turn is framed once per model request
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the turn is framed once per model request$`, thenFramedOncePerRequest)
	})
}

// thenFramedOncePerRequest (必查 呈現結果): the count of `╭─⠿ Turn N` frames
// equals the number of model requests the run made (round 034 FR-008).
func thenFramedOncePerRequest(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	want := modelRequestCount(sc)
	if want == 0 {
		return fmt.Errorf("the scenario recorded no model request (no fake request); stderr=%q", sc.stderr)
	}
	got := len(turnFrameNumbers(sc.stderr))
	if got != want {
		return fmt.Errorf("expected one frame per model request: got %d frames for %d requests; stderr=%q", got, want, sc.stderr)
	}
	return nil
}

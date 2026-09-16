package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T014 [BDD-RED] — Then: the run reports the post-turn status once per model request
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the run reports the post-turn status once per model request$`, thenPostTurnStatusPerRequest)
	})
}

// thenPostTurnStatusPerRequest (必查 呈現結果): the count of `╰─⠿ Ready` tails
// equals the number of model requests the run made — each call emits its own
// tail (round 034 FR-008).
func thenPostTurnStatusPerRequest(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	want := modelRequestCount(sc)
	if want == 0 {
		return fmt.Errorf("the scenario recorded no model request (no fake request); stderr=%q", sc.stderr)
	}
	got := readyLineCount(sc.stderr)
	if got != want {
		return fmt.Errorf("expected one post-turn status tail per model request: got %d tails for %d requests; stderr=%q", got, want, sc.stderr)
	}
	return nil
}

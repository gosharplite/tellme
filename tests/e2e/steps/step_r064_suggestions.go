package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// R064 — Given: the shared prompt log holds {count} newer prompts about other topics.
//
// Round 064 (ADR 0034): the depth-distinguishing fixture. It appends {count}
// unrelated, non-matching records to the user-global shared log so a
// previously-recorded matching prompt is pushed beyond the *newest 10* window
// but stays inside the newest-50 candidate pool — the only arrangement that can
// tell the deepened pool apart from the shallow one.
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the shared prompt log holds (\d+) newer prompts about other topics$`, givenSharedLogHoldsNewerUnrelated)
	})
}

func givenSharedLogHoldsNewerUnrelated(ctx context.Context, count int) error {
	sc := scenarioFrom(ctx)
	for i := 0; i < count; i++ {
		if err := appendPromptLog(sc, fmt.Sprintf("unrelated prompt %d", i+1)); err != nil {
			return err
		}
	}
	return nil
}

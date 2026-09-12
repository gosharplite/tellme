package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T014 — Then: tellme completes the turn without waiting for terminal input
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme completes the turn without waiting for terminal input$`, thenCompletesWithoutWaiting)
	})
}

// thenCompletesWithoutWaiting (必查 呈現結果): the run completed cleanly with the
// prompt read from a piped (non-terminal) standard input — i.e. the process did
// not block awaiting interactive input. Piped runs execute under a bounded
// deadline (harness.RunWithStdin), so a hang surfaces here as a non-nil runErr
// carrying an explicit deadline message rather than being inferred from the
// suite's own hang-to-timeout (grill Q6).
func thenCompletesWithoutWaiting(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if sc.runErr != nil {
		return fmt.Errorf("the run did not complete cleanly within the deadline (it may have waited): %v", sc.runErr)
	}
	if !sc.stdinSet {
		return fmt.Errorf("expected the prompt to arrive on piped standard input")
	}
	return nil
}

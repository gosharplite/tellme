package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T018 — Then: tellme stopped after {count} tool iterations
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme stopped after (\d+) tool iterations$`, thenStoppedAfter)
	})
}

// thenStoppedAfter (必查 權威狀態): the fake recorded exactly {count} tool-iteration
// rounds (one per request that carried a tool result) under a perpetually
// tool-requesting provider, matching the arranged MAX_TOOL_LOOP bound.
func thenStoppedAfter(ctx context.Context, count int) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	if got := countToolIterations(f); got != count {
		return fmt.Errorf("the run made %d tool iterations, want %d", got, count)
	}
	return nil
}

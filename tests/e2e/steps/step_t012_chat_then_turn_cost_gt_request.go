package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T012 — Then: the reported turn cost is greater than the request cost
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the reported turn cost is greater than the request cost$`, thenTurnCostGreater)
	})
}

// thenTurnCostGreater (必查 呈現結果): a tool-using turn's second `$` (the turn) is
// strictly greater than its first `$` (the just-returned call).
func thenTurnCostGreater(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	costs, _, _, _, _, ok := readyValues(sc.stderr)
	if !ok {
		return fmt.Errorf("standard error carried no Ready summary line; stderr=%q", sc.stderr)
	}
	if !(costs[1] > costs[0]) {
		return fmt.Errorf("turn cost %.4f is not greater than request cost %.4f", costs[1], costs[0])
	}
	return nil
}

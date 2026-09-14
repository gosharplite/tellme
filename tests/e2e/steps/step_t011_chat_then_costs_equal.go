package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T011 — Then: the reported request, turn, and session costs are equal
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the reported request, turn, and session costs are equal$`, thenCostsEqual)
	})
}

// thenCostsEqual (必查 呈現結果): a fresh session reports request == turn == session.
func thenCostsEqual(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	costs, _, _, _, _, ok := readyValues(sc.stderr)
	if !ok {
		return fmt.Errorf("standard error carried no Ready summary line; stderr=%q", sc.stderr)
	}
	if costs[0] != costs[1] || costs[1] != costs[2] {
		return fmt.Errorf("costs are not equal: request=%.4f turn=%.4f session=%.4f", costs[0], costs[1], costs[2])
	}
	return nil
}

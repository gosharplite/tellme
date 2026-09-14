package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T016 — Then: the run reports a zero cost
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the run reports a zero cost$`, thenZeroCost)
	})
}

// thenZeroCost (必查 呈現結果): the summary line's `$` costs are `$0.0000` (the
// model has no `MODELS` pricing entry — config-only pricing, research D2).
func thenZeroCost(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	costs, _, _, _, _, ok := readyValues(sc.stderr)
	if !ok {
		return fmt.Errorf("standard error carried no Ready summary line; stderr=%q", sc.stderr)
	}
	for i, c := range costs {
		if c != 0 {
			return fmt.Errorf("cost #%d is %.4f, want $0.0000 for an un-priced model", i+1, c)
		}
	}
	return nil
}

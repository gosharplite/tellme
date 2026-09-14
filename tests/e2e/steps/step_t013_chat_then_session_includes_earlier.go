package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T013 — Then: the reported session summary includes the earlier call
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the reported session summary includes the earlier call$`, thenSessionIncludesEarlier)
	})
}

// thenSessionIncludesEarlier (必查 呈現結果): the session (`$` #3) and its token
// totals include the arranged prior call — strictly greater than the current
// call alone.
func thenSessionIncludesEarlier(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	costs, _, _, _, _, ok := readyValues(sc.stderr)
	if !ok {
		return fmt.Errorf("standard error carried no Ready summary line; stderr=%q", sc.stderr)
	}
	if !(costs[2] > costs[0]) {
		return fmt.Errorf("session cost %.4f does not include the earlier call (request cost %.4f)", costs[2], costs[0])
	}
	return nil
}

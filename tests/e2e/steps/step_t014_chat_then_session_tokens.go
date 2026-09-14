package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T014 — Then: the run reports the session's missed, cached, and output tokens
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the run reports the session's missed, cached, and output tokens$`, thenSessionTokens)
	})
}

// thenSessionTokens (必查 呈現結果): the summary line carries the session token
// totals `M: <n> H: <n> O: <n>`.
func thenSessionTokens(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	_, _, _, _, _, ok := readyValues(sc.stderr)
	if !ok {
		return fmt.Errorf("standard error carried no Ready summary line with session token totals; stderr=%q", sc.stderr)
	}
	return nil
}

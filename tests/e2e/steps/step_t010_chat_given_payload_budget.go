package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T010 — Given: the payload budget is "{budget}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the payload budget is "([^"]*)"$`, givenPayloadBudget)
	})
}

// givenPayloadBudget sets MAX_HISTORY_TOKENS in the subprocess environment
// (怎麼做 / 權威狀態落地: the effective payload budget resolves to {budget}).
func givenPayloadBudget(ctx context.Context, budget string) error {
	sc := scenarioFrom(ctx)
	sc.setEnv("MAX_HISTORY_TOKENS", budget)
	return nil
}

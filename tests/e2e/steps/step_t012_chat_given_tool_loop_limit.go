package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T012 — Given: the tool-loop limit is "{limit}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the tool-loop limit is "([^"]*)"$`, givenToolLoopLimit)
	})
}

// givenToolLoopLimit (怎麼做 / 權威狀態落地 / 回寫): set the MAX_TOOL_LOOP
// environment override for the next run so the effective loop bound resolves to
// {limit}.
func givenToolLoopLimit(ctx context.Context, limit string) error {
	sc := scenarioFrom(ctx)
	sc.setEnv("MAX_TOOL_LOOP", limit)
	return nil
}

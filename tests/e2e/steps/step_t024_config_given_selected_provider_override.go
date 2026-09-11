package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T024 — Given: the selected provider override is "{provider}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the selected provider override is "([^"]*)"$`, givenSelectedProviderOverride)
	})
}

// givenSelectedProviderOverride sets TELL_ME_SELECTED_PROVIDER (怎麼做 /
// 權威狀態落地: the effective selected provider is taken from the environment
// first).
func givenSelectedProviderOverride(ctx context.Context, provider string) error {
	scenarioFrom(ctx).setEnv("TELL_ME_SELECTED_PROVIDER", provider)
	return nil
}

package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T021 — Given: the effective mode is "{mode}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the effective mode is "([^"]*)"$`, givenEffectiveMode)
	})
}

// givenEffectiveMode sets TELL_ME_MODE (怎麼做 / 權威狀態落地: the effective mode
// resolves from the environment first).
func givenEffectiveMode(ctx context.Context, mode string) error {
	scenarioFrom(ctx).setEnv("TELL_ME_MODE", mode)
	return nil
}

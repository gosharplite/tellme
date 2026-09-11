package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T031 — Given: the mode override is "{mode}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the mode override is "([^"]*)"$`, givenModeOverride)
	})
}

// givenModeOverride sets TELL_ME_MODE (怎麼做 / 權威狀態落地: the effective mode
// is taken from the environment first).
func givenModeOverride(ctx context.Context, mode string) error {
	scenarioFrom(ctx).setEnv("TELL_ME_MODE", mode)
	return nil
}

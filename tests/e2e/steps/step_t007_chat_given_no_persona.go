package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T007 — Given: the runtime home holds a configuration with no persona
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the runtime home holds a configuration with no persona$`, givenNoPersonaConfig)
	})
}

// givenNoPersonaConfig clears the default configuration's PERSON (怎麼做 /
// 權威狀態落地 / 回寫). The default configuration must already exist.
func givenNoPersonaConfig(ctx context.Context) error {
	return scenarioFrom(ctx).setPersona("")
}

package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T006 — Given: the runtime home holds a configuration whose persona is "{persona}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the runtime home holds a configuration whose persona is "([^"]*)"$`, givenPersonaConfig)
	})
}

// givenPersonaConfig sets the default configuration's PERSON to {persona}
// (怎麼做 / 權威狀態落地 / 回寫). The default configuration must already exist
// (arranged by the provider Given).
func givenPersonaConfig(ctx context.Context, persona string) error {
	return scenarioFrom(ctx).setPersona(unescapeText(persona))
}

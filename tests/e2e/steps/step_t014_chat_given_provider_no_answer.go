package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T014 — Given: a configured provider "{provider}" whose endpoint answers with no usable answer
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint answers with no usable answer$`, givenConfiguredProviderNoAnswer)
	})
}

// givenConfiguredProviderNoAnswer writes a resolvable default configuration
// selecting {provider} whose endpoint points at the fake, and scripts the fake
// to answer with a body carrying no usable answer (怎麼做 / 權威狀態落地 / 回寫).
func givenConfiguredProviderNoAnswer(ctx context.Context, provider string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	f.NoAnswer()
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

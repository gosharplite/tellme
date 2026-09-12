package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T012 — Given: a configured provider "{provider}" whose endpoint answers with an error status
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint answers with an error status$`, givenConfiguredProviderErrorStatus)
	})
}

// givenConfiguredProviderErrorStatus writes a resolvable default configuration
// selecting {provider} whose endpoint points at the fake, and scripts the fake
// to answer with a non-2xx status (怎麼做 / 權威狀態落地 / 回寫).
func givenConfiguredProviderErrorStatus(ctx context.Context, provider string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	f.ErrorStatus(500)
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

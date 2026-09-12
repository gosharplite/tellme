package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T013 — Given: a configured provider "{provider}" whose endpoint is unreachable
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint is unreachable$`, givenConfiguredProviderUnreachable)
	})
}

// givenConfiguredProviderUnreachable writes a resolvable default configuration
// selecting {provider} whose endpoint points at a listener that has been closed,
// so the connection is refused (怎麼做 / 權威狀態落地 / 回寫).
func givenConfiguredProviderUnreachable(ctx context.Context, provider string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	url := f.URL()
	f.Close() // release the listener so the endpoint refuses connections
	return sc.writeDefaultConfig(provider, map[string]string{provider: url})
}

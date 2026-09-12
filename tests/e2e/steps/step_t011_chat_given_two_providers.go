package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T011 — Given: configured providers "{provider_a}" and "{provider_b}" whose endpoints answer
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^configured providers "([^"]*)" and "([^"]*)" whose endpoints answer$`, givenTwoConfiguredProviders)
	})
}

// givenTwoConfiguredProviders writes a resolvable default configuration with two
// providers, each backed by its own fake (怎麼做 / 權威狀態落地 / 回寫).
func givenTwoConfiguredProviders(ctx context.Context, providerA, providerB string) error {
	sc := scenarioFrom(ctx)
	fa := sc.newFake()
	fa.Answer(providerA + " answer")
	fb := sc.newFake()
	fb.Answer(providerB + " answer")
	sc.registerFake(providerA, fa)
	sc.registerFake(providerB, fb)
	return sc.writeDefaultConfig(providerA, map[string]string{providerA: fa.URL(), providerB: fb.URL()})
}

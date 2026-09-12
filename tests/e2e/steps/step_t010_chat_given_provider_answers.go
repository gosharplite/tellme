package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T010 — Given: a configured provider "{provider}" whose endpoint answers with "{answer}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint answers with "([^"]*)"$`, givenConfiguredProviderAnswers)
	})
}

// givenConfiguredProviderAnswers writes a resolvable default configuration
// selecting {provider} whose endpoint points at the fake provider, and scripts
// the fake to answer {answer} (怎麼做 / 權威狀態落地 / 回寫).
func givenConfiguredProviderAnswers(ctx context.Context, provider, answer string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	f.Answer(answer)
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

package steps

import (
	"context"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// T016 [BDD-RED] — Given: a configured provider "{provider}" whose endpoint reports the offered tools and then answers with "{answer}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint reports the offered tools and then answers with "([^"]*)"$`, givenProviderReportsTools)
	})
}

// givenProviderReportsTools (怎麼做 / 權威狀態落地 / 回寫): script the fake to answer
// {answer} (the fake records the request's offered tool definitions, which the
// offered-tools Then asserts on).
func givenProviderReportsTools(ctx context.Context, provider, answer string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	f.Script(fakeprovider.Reply{Answer: unescapeText(answer)})
	sc.scriptedAnswer = unescapeText(answer)
	sc.scriptedAnswerSet = true
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

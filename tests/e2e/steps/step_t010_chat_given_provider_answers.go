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
// the fake to answer {answer} (怎麼做 / 權威狀態落地 / 回寫). The answer is decoded
// through the `\n`/`\t`/`\\`/`\xHH` escape convention so a newline- or
// control-byte-bearing answer is expressible (grill Q6); the decoded value is
// recorded on the scenario for the decoration check.
func givenConfiguredProviderAnswers(ctx context.Context, provider, answer string) error {
	sc := scenarioFrom(ctx)
	answer = unescapeText(answer)
	f := sc.newFake()
	f.Answer(answer)
	sc.scriptedAnswer = answer
	sc.scriptedAnswerSet = true
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

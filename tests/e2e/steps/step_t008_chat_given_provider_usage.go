package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T008 — Given: a configured provider "{provider}" whose endpoint answers with "{answer}" and reports its usage
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint answers with "([^"]*)" and reports its usage$`, givenProviderReportsUsage)
	})
}

// givenProviderReportsUsage writes a resolvable default configuration selecting
// {provider} pointing at the fake, scripts the fake to answer {answer}, and
// scripts a `usage` block (怎麼做 / 權威狀態落地 / 回寫).
func givenProviderReportsUsage(ctx context.Context, provider, answer string) error {
	sc := scenarioFrom(ctx)
	answer = unescapeText(answer)
	f := sc.newFake()
	f.Answer(answer)
	f.ReportUsage(137, 9)
	sc.scriptedAnswer = answer
	sc.scriptedAnswerSet = true
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

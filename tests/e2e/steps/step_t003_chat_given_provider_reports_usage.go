package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T003 — Given: a configured provider "{provider}" whose endpoint answers with "{answer}" and reports the token usage:
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint answers with "([^"]*)" and reports the token usage:$`, givenProviderReportsTokenUsage)
	})
}

// givenProviderReportsTokenUsage (怎麼做 / 權威狀態落地 / 回寫): write the resolvable
// default config selecting {provider}; script the fake to answer {answer} with a
// JSON `usage` block whose cited counts are the (inclusive) wire value — the
// DataTable's `| prompt | cached | completion | thinking |` row.
func givenProviderReportsTokenUsage(ctx context.Context, provider, answer string, table *godog.Table) error {
	sc := scenarioFrom(ctx)
	answer = unescapeText(answer)
	prompt, cached, completion, thinking, err := usageFromTable(table)
	if err != nil {
		return err
	}
	f := sc.newFake()
	f.Answer(answer)
	f.ReportUsageDetails(prompt, cached, completion, thinking)
	sc.scriptedAnswer = answer
	sc.scriptedAnswerSet = true
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

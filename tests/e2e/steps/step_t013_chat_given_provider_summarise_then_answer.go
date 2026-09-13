package steps

import (
	"context"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// T013 — Given: a configured provider "{provider}" whose endpoint asks tellme to summarise the conversation and then answers with "{answer}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint asks tellme to summarise the conversation and then answers with "([^"]*)"$`, givenProviderSummariseThenAnswer)
	})
}

// givenProviderSummariseThenAnswer (怎麼做 / 權威狀態落地 / 回寫): write the config
// selecting {provider}; script the fake to return a `summarize_history` tool-call
// on the first request, then a final answer {answer}.
func givenProviderSummariseThenAnswer(ctx context.Context, provider, answer string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	f.Script(
		fakeprovider.Reply{ToolName: "summarize_history", Arguments: "{}"},
		fakeprovider.Reply{Answer: unescapeText(answer)},
	)
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

package steps

import (
	"context"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// T004 [BDD-RED] — Given: a configured provider "{provider}" whose endpoint asks
// tellme to read "{path_a}" and "{path_b}" and then answers with "{answer}".
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint asks tellme to read "([^"]*)" and "([^"]*)" and then answers with "([^"]*)"$`, givenProviderReadTwoThenAnswer)
	})
}

// givenProviderReadTwoThenAnswer (怎麼做 / 權威狀態落地 / 回寫): write the resolvable
// default config selecting {provider}; script the fake to return ONE response
// carrying two `read_files` tool calls ({path_a}, {path_b}), then a final answer
// {answer} on the next request (a one-round, two-tool exchange).
func givenProviderReadTwoThenAnswer(ctx context.Context, provider, pathA, pathB, answer string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	f.Script(
		fakeprovider.Reply{Tools: []fakeprovider.ToolRequest{
			{Name: "read_files", Arguments: readArgs(pathA)},
			{Name: "read_files", Arguments: readArgs(pathB)},
		}},
		fakeprovider.Reply{Answer: unescapeText(answer)},
	)
	sc.scriptedAnswer = unescapeText(answer)
	sc.scriptedAnswerSet = true
	sc.scriptedTool = "read_files"
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

package steps

import (
	"context"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// T015 [BDD-RED] — Given: a configured provider "{provider}" whose endpoint shows the folder tree and then answers with "{answer}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint shows the folder tree and then answers with "([^"]*)"$`, givenProviderShowsTree)
	})
}

// givenProviderShowsTree (怎麼做 / 權威狀態落地 / 回寫): script the fake to return a
// get_tree tool call (reason required; path/depth default), then the answer.
func givenProviderShowsTree(ctx context.Context, provider, answer string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	f.Script(
		fakeprovider.Reply{ToolName: "get_tree", Arguments: `{"reason":"show the folder tree"}`},
		fakeprovider.Reply{Answer: unescapeText(answer)},
	)
	sc.scriptedAnswer = unescapeText(answer)
	sc.scriptedAnswerSet = true
	sc.scriptedTool = "get_tree"
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

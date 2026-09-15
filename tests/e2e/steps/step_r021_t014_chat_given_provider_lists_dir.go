package steps

import (
	"context"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// T014 [BDD-RED] — Given: a configured provider "{provider}" whose endpoint lists the current directory and then answers with "{answer}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint lists the current directory and then answers with "([^"]*)"$`, givenProviderListsDir)
	})
}

// givenProviderListsDir (怎麼做 / 權威狀態落地 / 回寫): script the fake to return a
// list_files tool call (no path — it defaults to "."), then the answer {answer}.
func givenProviderListsDir(ctx context.Context, provider, answer string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	f.Script(
		fakeprovider.Reply{ToolName: "list_files", Arguments: `{"reason":"list the working directory"}`},
		fakeprovider.Reply{Answer: unescapeText(answer)},
	)
	sc.scriptedAnswer = unescapeText(answer)
	sc.scriptedAnswerSet = true
	sc.scriptedTool = "list_files"
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

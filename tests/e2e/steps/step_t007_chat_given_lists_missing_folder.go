package steps

import (
	"context"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// T007 [BDD-RED] — Given: a configured provider "{provider}" whose endpoint lists a folder that does not exist and then answers with "{answer}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint lists a folder that does not exist and then answers with "([^"]*)"$`, givenProviderListsMissingFolder)
	})
}

// givenProviderListsMissingFolder (怎麼做 / 權威狀態落地 / 回寫): write the
// resolvable default config selecting {provider}; script the fake to return a
// `list_files` tool call for a non-existent path (so the tool returns a non-nil
// error), then a final answer {answer}.
func givenProviderListsMissingFolder(ctx context.Context, provider, answer string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	f.Script(
		fakeprovider.Reply{ToolName: "list_files", Arguments: `{"path":"./this-folder-does-not-exist","reason":"list a missing folder"}`},
		fakeprovider.Reply{Answer: unescapeText(answer)},
	)
	sc.scriptedAnswer = unescapeText(answer)
	sc.scriptedAnswerSet = true
	sc.scriptedTool = "list_files"
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

package steps

import (
	"context"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// T011 [BDD-RED] — Given: a configured provider "{provider}" whose endpoint asks tellme to read "{path_a}" and "{path_b}" in one request and then answers with "{answer}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint asks tellme to read "([^"]*)" and "([^"]*)" in one request and then answers with "([^"]*)"$`, givenProviderReadTwoOneRequest)
	})
}

// givenProviderReadTwoOneRequest (怎麼做 / 權威狀態落地 / 回寫): script the fake to
// return ONE response carrying a single read_files tool call whose `filepaths` are
// {path_a} and {path_b}, then the answer {answer}.
func givenProviderReadTwoOneRequest(ctx context.Context, provider, pathA, pathB, answer string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	f.Script(
		fakeprovider.Reply{ToolName: "read_files", Arguments: readFilesArgs([]string{pathA, pathB})},
		fakeprovider.Reply{Answer: unescapeText(answer)},
	)
	sc.scriptedAnswer = unescapeText(answer)
	sc.scriptedAnswerSet = true
	sc.scriptedTool = "read_files"
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

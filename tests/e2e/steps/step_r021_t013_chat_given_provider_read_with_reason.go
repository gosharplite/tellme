package steps

import (
	"context"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// T013 [BDD-RED] — Given: a configured provider "{provider}" whose endpoint asks tellme to read "{path}" with the reason "{reason}" and then answers with "{answer}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint asks tellme to read "([^"]*)" with the reason "([^"]*)" and then answers with "([^"]*)"$`, givenProviderReadWithReason)
	})
}

// givenProviderReadWithReason (怎麼做 / 權威狀態落地 / 回寫): script the fake to
// return a read_files tool call carrying `filepaths` = [{path}] and a top-level
// `reason` = {reason}, then the answer {answer}; the loop must echo the reason.
func givenProviderReadWithReason(ctx context.Context, provider, path, reason, answer string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	f.Script(
		fakeprovider.Reply{ToolName: "read_files", Arguments: readArgsWithReason(path, reason)},
		fakeprovider.Reply{Answer: unescapeText(answer)},
	)
	sc.scriptedAnswer = unescapeText(answer)
	sc.scriptedAnswerSet = true
	sc.scriptedTool = "read_files"
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

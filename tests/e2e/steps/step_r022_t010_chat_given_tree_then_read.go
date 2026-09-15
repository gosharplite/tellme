package steps

import (
	"context"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// T010 [BDD-RED] — Given: a configured provider "{provider}" whose endpoint shows the folder tree and then reads "{path}" and then answers with "{answer}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint shows the folder tree and then reads "([^"]*)" and then answers with "([^"]*)"$`, givenProviderTreeThenRead)
	})
}

// givenProviderTreeThenRead (怎麼做 / 權威狀態落地 / 回寫): script the fake for a
// three-step exchange — a `get_tree` call, then a `read_files` call for {path},
// then the answer {answer} — so the loop reports the two tools in call order
// (round-022 FR-005 order carrier, review TD3).
func givenProviderTreeThenRead(ctx context.Context, provider, path, answer string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	f.Script(
		fakeprovider.Reply{ToolName: "get_tree", Arguments: `{"reason":"survey the project"}`},
		fakeprovider.Reply{ToolName: "read_files", Arguments: readArgs(path)},
		fakeprovider.Reply{Answer: unescapeText(answer)},
	)
	sc.scriptedAnswer = unescapeText(answer)
	sc.scriptedAnswerSet = true
	sc.scriptedTool = "get_tree"
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

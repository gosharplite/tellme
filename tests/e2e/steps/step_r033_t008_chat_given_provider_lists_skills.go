package steps

import (
	"context"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// T008 [BDD-RED] — Given: a configured provider "{provider}" whose endpoint asks tellme to list its skills and then answers with "{answer}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint asks tellme to list its skills and then answers with "([^"]*)"$`, givenProviderListsSkills)
	})
}

// givenProviderListsSkills (怎麼做 / 權威狀態落地 / 回寫): write the configuration and
// script the fake to return a `list_skills` tool-call response (no arguments
// beyond the required `reason`), then a final answer {answer}.
func givenProviderListsSkills(ctx context.Context, provider, answer string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	f.Script(
		fakeprovider.Reply{ToolName: "list_skills", Arguments: `{"reason":"list the available skills"}`},
		fakeprovider.Reply{Answer: unescapeText(answer)},
	)
	sc.scriptedAnswer = unescapeText(answer)
	sc.scriptedAnswerSet = true
	sc.scriptedTool = "list_skills"
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

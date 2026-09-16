package steps

import (
	"context"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// T009 [BDD-RED] — Given: a configured provider "{provider}" whose endpoint asks tellme to list its skills, then to read the skill "{name}", and then answers with "{answer}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint asks tellme to list its skills, then to read the skill "([^"]*)", and then answers with "([^"]*)"$`, givenProviderListsThenReadsSkill)
	})
}

// givenProviderListsThenReadsSkill (怎麼做 / 權威狀態落地 / 回寫): script the fake for a
// three-step exchange — a `list_skills` call, then a `read_files` call for the
// skill {name}'s file (the path the listing exposes), then a final answer
// {answer}.
func givenProviderListsThenReadsSkill(ctx context.Context, provider, name, answer string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	f.Script(
		fakeprovider.Reply{ToolName: "list_skills", Arguments: `{"reason":"list the available skills"}`},
		fakeprovider.Reply{ToolName: "read_files", Arguments: readArgs(sc.homePath(skillRelPath(name)))},
		fakeprovider.Reply{Answer: unescapeText(answer)},
	)
	sc.scriptedAnswer = unescapeText(answer)
	sc.scriptedAnswerSet = true
	sc.scriptedTool = "read_files"
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

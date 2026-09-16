package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T005 [BDD-RED] — Given: the runtime home holds a skill "{name}" described as "{description}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the runtime home holds a skill "([^"]*)" described as "([^"]*)"$`, givenRuntimeHomeHoldsSkill)
	})
}

// givenRuntimeHomeHoldsSkill (怎麼做 / 權威狀態落地 / 回寫): create
// $TELL_ME_HOME/docs/skills/{name}/SKILL.md whose content is a YAML frontmatter
// block declaring {name} / {description} followed by a Markdown body.
func givenRuntimeHomeHoldsSkill(ctx context.Context, name, description string) error {
	sc := scenarioFrom(ctx)
	return sc.writeFile(skillRelPath(name), []byte(skillFileMarkdown(name, description)))
}

package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T006 [BDD-RED] — Given: the runtime home holds a skill "{name}" whose folder also holds a reference file "{relpath}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the runtime home holds a skill "([^"]*)" whose folder also holds a reference file "([^"]*)"$`, givenRuntimeHomeHoldsSkillWithReference)
	})
}

// givenRuntimeHomeHoldsSkillWithReference (怎麼做 / 權威狀態落地 / 回寫): create the
// skill {name} (frontmatter name/description) at docs/skills/{name}/SKILL.md, and
// a reference file at docs/skills/{name}/{relpath} whose content has NO
// frontmatter block — so it is not a skill.
func givenRuntimeHomeHoldsSkillWithReference(ctx context.Context, name, relpath string) error {
	sc := scenarioFrom(ctx)
	if err := sc.writeFile(skillRelPath(name), []byte(skillFileMarkdown(name, "a skill with a reference file"))); err != nil {
		return err
	}
	return sc.writeFile(skillReferenceRelPath(name, relpath), []byte("# Reference\n\nThis is non-skill reference material.\n"))
}

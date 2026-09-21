package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// Round-075 Given: author a skill whose `description` is a YAML folded block
// scalar. Kept in its own file so the per-sentence step files stay independent
// (Zero Shared Edits).
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the runtime home holds a skill "([^"]*)" whose description is written as a folded block scalar across the lines "([^"]*)" and "([^"]*)"$`, givenRuntimeHomeHoldsFoldedSkill)
	})
}

// foldedSkillMarkdown renders a SKILL.md whose frontmatter declares `name`
// inline and `description` as a folded block scalar (`description: >` then two
// indented lines). A conforming reader folds the two lines to a single space.
func foldedSkillMarkdown(name, first, second string) string {
	return "---\nname: " + name + "\ndescription: >\n  " + first + "\n  " + second + "\n---\n# " + name + "\n"
}

// givenRuntimeHomeHoldsFoldedSkill (怎麼做 / 權威狀態落地 / 回寫): create
// $TELL_ME_HOME/docs/skills/{name}/SKILL.md with a folded block-scalar description.
func givenRuntimeHomeHoldsFoldedSkill(ctx context.Context, name, first, second string) error {
	sc := scenarioFrom(ctx)
	return sc.writeFile(skillRelPath(name), []byte(foldedSkillMarkdown(name, first, second)))
}

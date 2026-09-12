package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T007 — When: the operator pipes "{content}" into tellme with the instruction "{instruction}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.When(`^the operator pipes "([^"]*)" into tellme with the instruction "([^"]*)"$`, whenPipesWithInstruction)
	})
}

// whenPipesWithInstruction runs `tellme "{instruction}"` with {content} on a
// piped standard input (怎麼做 / 權威狀態落地 / 回寫).
func whenPipesWithInstruction(ctx context.Context, content, instruction string) error {
	sc := scenarioFrom(ctx)
	sc.pipeStdin(content)
	sc.args = []string{instruction}
	sc.run()
	return nil
}

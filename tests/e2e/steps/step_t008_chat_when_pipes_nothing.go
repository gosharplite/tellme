package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T008 — When: the operator pipes nothing into tellme
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.When(`^the operator pipes nothing into tellme$`, whenPipesNothing)
	})
}

// whenPipesNothing runs `tellme` (no positional prompt) with an EMPTY piped
// standard input (immediate EOF) (怎麼做 / 權威狀態落地 / 回寫).
func whenPipesNothing(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	sc.pipeStdin("")
	sc.args = nil
	sc.run()
	return nil
}

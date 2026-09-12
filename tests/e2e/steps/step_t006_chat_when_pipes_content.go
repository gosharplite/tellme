package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T006 — When: the operator pipes "{content}" into tellme
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.When(`^the operator pipes "([^"]*)" into tellme$`, whenPipesContent)
	})
}

// whenPipesContent runs `tellme` (no positional prompt) with {content} on a
// piped standard input (怎麼做 / 權威狀態落地 / 回寫: run to completion; capture
// exit code, stdout, stderr, and the fake's recorded requests).
func whenPipesContent(ctx context.Context, content string) error {
	sc := scenarioFrom(ctx)
	sc.pipeStdin(content)
	sc.args = nil
	sc.run()
	return nil
}

package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// Round-012 amendment A8 — When: the operator starts a fresh session with "--new"
// and pipes "{content}".
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.When(`^the operator starts a fresh session with "--new" and pipes "([^"]*)"$`, whenNewPipesContent)
	})
}

// whenNewPipesContent runs `tellme --new` (no positional prompt) with {content}
// on a piped standard input (怎麼做 / 權威狀態落地 / 回寫: run to completion; capture
// exit code, stdout, stderr, and the fake's recorded requests). Combined with the
// forced-terminal Given, the CLI archives the session and then reads {content} at
// the "terminal".
func whenNewPipesContent(ctx context.Context, content string) error {
	sc := scenarioFrom(ctx)
	sc.pipeStdin(content)
	sc.lastPrompt = content
	sc.args = []string{"--new"}
	sc.run()
	return nil
}

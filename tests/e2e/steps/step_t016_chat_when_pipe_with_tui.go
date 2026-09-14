package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T016 — When: the operator pipes "{content}" into tellme with the interactive prompt enabled
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.When(`^the operator pipes "([^"]*)" into tellme with the interactive prompt enabled$`, whenPipeWithTUI)
	})
}

// whenPipeWithTUI runs `tellme -i` with {content} on a piped standard input — a
// NON-terminal, so the interactive prompt must not engage and the piped path
// applies (怎麼做 / 權威狀態落地 / 回寫). The forced-terminal seam is deliberately
// NOT set here.
func whenPipeWithTUI(ctx context.Context, content string) error {
	sc := scenarioFrom(ctx)
	sc.pipeStdin(content)
	sc.args = []string{"-i"}
	sc.run()
	return nil
}

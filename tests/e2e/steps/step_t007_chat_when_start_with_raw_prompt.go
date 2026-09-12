package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T007 — When: the operator starts tellme with the prompt "{prompt}" and the raw flag
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.When(`^the operator starts tellme with the prompt "([^"]*)" and the raw flag$`, whenStartsWithRawPrompt)
	})
}

// whenStartsWithRawPrompt runs `tellme -r "{prompt}"` (no -c) under the current
// environment; the raw path bypasses rendering (怎麼做 / 權威狀態落地 / 回寫: run to
// completion; capture exit code, stdout, stderr, and the fake's recorded
// requests).
func whenStartsWithRawPrompt(ctx context.Context, prompt string) error {
	sc := scenarioFrom(ctx)
	sc.lastPrompt = prompt
	sc.args = []string{"-r", prompt}
	sc.run()
	return nil
}

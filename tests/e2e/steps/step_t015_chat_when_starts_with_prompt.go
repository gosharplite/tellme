package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T015 — When: the operator starts tellme with the prompt "{prompt}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.When(`^the operator starts tellme with the prompt "([^"]*)"$`, whenStartsWithPrompt)
	})
}

// whenStartsWithPrompt runs `tellme "{prompt}"` (no -c) under the current
// environment (怎麼做 / 權威狀態落地 / 回寫: run to completion; capture exit
// code, stdout, stderr, and the fake's recorded requests).
func whenStartsWithPrompt(ctx context.Context, prompt string) error {
	sc := scenarioFrom(ctx)
	sc.lastPrompt = prompt
	sc.args = []string{prompt}
	sc.run()
	return nil
}

package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T007 — When: the operator starts a fresh session with "--new" and the prompt "{prompt}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.When(`^the operator starts a fresh session with "--new" and the prompt "([^"]*)"$`, whenStartsFreshWithPrompt)
	})
}

// whenStartsFreshWithPrompt (怎麼做 / 權威狀態落地 / 回寫): run `tellme --new
// "{prompt}"` under the current environment and capture the result.
func whenStartsFreshWithPrompt(ctx context.Context, prompt string) error {
	sc := scenarioFrom(ctx)
	sc.args = []string{"--new", unescapeText(prompt)}
	sc.run()
	return nil
}

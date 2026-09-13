package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T009 — When: the operator uses the persona "{persona}" and starts tellme with the prompt "{prompt}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.When(`^the operator uses the persona "([^"]*)" and starts tellme with the prompt "([^"]*)"$`, whenPersonaRun)
	})
}

// whenPersonaRun sets the default configuration's persona and runs
// `tellme "{prompt}"` (怎麼做 / 權威狀態落地 / 回寫) — a single Act.
func whenPersonaRun(ctx context.Context, persona, prompt string) error {
	sc := scenarioFrom(ctx)
	if err := sc.setPersona(unescapeText(persona)); err != nil {
		return err
	}
	sc.lastPrompt = prompt
	sc.args = []string{prompt}
	sc.run()
	return nil
}

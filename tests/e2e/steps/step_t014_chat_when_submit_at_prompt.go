package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T014 — When: the operator submits the prompt "{prompt}" at the interactive prompt
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.When(`^the operator submits the prompt "([^"]*)" at the interactive prompt$`, whenSubmitAtPrompt)
	})
}

// whenSubmitAtPrompt runs `tellme -i` against the forced-terminal seam with a
// scripted sequence (open → type {prompt} → submit), so the submitted text
// becomes the prompt of one reasoning turn (怎麼做 / 權威狀態落地 / 回寫).
func whenSubmitAtPrompt(ctx context.Context, prompt string) error {
	sc := scenarioFrom(ctx)
	sc.lastPrompt = prompt
	launchTUI(sc, tuiKeysTypeAndSubmit(unescapeText(prompt)))
	return nil
}

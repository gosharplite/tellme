package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// R027 — When: the operator starts a fresh session at the interactive prompt and submits the prompt "{prompt}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.When(`^the operator starts a fresh session at the interactive prompt and submits the prompt "([^"]*)"$`, whenFreshSubmitAtPrompt)
	})
}

// whenFreshSubmitAtPrompt runs `tellme --new -i` against the forced-terminal seam
// with a scripted sequence (open → type {prompt} → submit), so the fresh session
// archives the active history before the submitted turn becomes the turn's prompt
// (怎麼做 / 權威狀態落地 / 回寫).
func whenFreshSubmitAtPrompt(ctx context.Context, prompt string) error {
	sc := scenarioFrom(ctx)
	sc.lastPrompt = prompt
	launchTUIFresh(sc, tuiKeysTypeAndSubmit(unescapeText(prompt)))
	return nil
}

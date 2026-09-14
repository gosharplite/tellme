package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T012 — When: the operator opens the interactive prompt
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.When(`^the operator opens the interactive prompt$`, whenOpenInteractivePrompt)
	})
}

// whenOpenInteractivePrompt runs `tellme -i` against the forced-terminal seam
// with a scripted key sequence (open → abort), so the prompt is engaged and the
// process exits (怎麼做 / 權威狀態落地 / 回寫).
func whenOpenInteractivePrompt(ctx context.Context) error {
	launchTUI(scenarioFrom(ctx), tuiKeysOpenAndAbort())
	return nil
}

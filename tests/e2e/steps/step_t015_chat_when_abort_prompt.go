package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T015 — When: the operator aborts the interactive prompt
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.When(`^the operator aborts the interactive prompt$`, whenAbortInteractivePrompt)
	})
}

// whenAbortInteractivePrompt runs `tellme -i` against the forced-terminal seam
// with a scripted sequence (open → abort), so the prompt is engaged and the
// process exits with no prompt sent (怎麼做 / 權威狀態落地 / 回寫).
func whenAbortInteractivePrompt(ctx context.Context) error {
	launchTUI(scenarioFrom(ctx), tuiKeysOpenAndAbort())
	return nil
}

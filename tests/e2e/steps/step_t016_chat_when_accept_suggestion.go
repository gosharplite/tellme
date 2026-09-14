package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T016 — When: the operator opens the interactive prompt, types "{query}", and accepts the current suggestion
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.When(`^the operator opens the interactive prompt, types "([^"]*)", and accepts the current suggestion$`, whenOpenTypeAndAccept)
	})
}

// whenOpenTypeAndAccept runs `tellme -i` against the forced-terminal seam with a
// scripted sequence (open → type {query} → Tab accept → abort), so the accepted
// suggestion is inserted into the editor (怎麼做 / 權威狀態落地 / 回寫).
func whenOpenTypeAndAccept(ctx context.Context, query string) error {
	launchTUI(scenarioFrom(ctx), tuiKeysTypeAcceptAbort(unescapeText(query)))
	return nil
}

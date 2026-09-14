package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T013 — When: the operator opens the interactive prompt and types "{query}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.When(`^the operator opens the interactive prompt and types "([^"]*)"$`, whenOpenAndType)
	})
}

// whenOpenAndType runs `tellme -i` against the forced-terminal seam with a
// scripted sequence (open → type {query} → abort), so the prompt computes
// suggestions for {query} and exits (怎麼做 / 權威狀態落地 / 回寫).
func whenOpenAndType(ctx context.Context, query string) error {
	launchTUI(scenarioFrom(ctx), tuiKeysTypeAndAbort(unescapeText(query)))
	return nil
}

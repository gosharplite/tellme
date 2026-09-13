package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T009 — Given: the configured Gemini provider entry allows at most {tokens} output tokens
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the configured Gemini provider entry allows at most (\d+) output tokens$`, givenGeminiProviderMaxTokens)
	})
}

// givenGeminiProviderMaxTokens sets the default configuration's selected provider
// entry MAX_TOKENS to {tokens} (怎麼做 / 回寫).
func givenGeminiProviderMaxTokens(ctx context.Context, tokens int) error {
	sc := scenarioFrom(ctx)
	return sc.setSelectedProviderMaxTokens(tokens)
}

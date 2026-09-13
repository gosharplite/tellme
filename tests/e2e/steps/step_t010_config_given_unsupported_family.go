package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T010 — Given: a configured provider "{provider}" of a family tellme cannot drive
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" of a family tellme cannot drive$`, givenProviderUnsupportedFamily)
	})
}

// givenProviderUnsupportedFamily writes a resolvable default configuration
// selecting {provider} whose entry TYPE is a family the transport cannot drive
// (e.g. anthropic) — 構造 / 權威狀態落地 (the config resolves; the family is
// unsupported).
func givenProviderUnsupportedFamily(ctx context.Context, provider string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	sc.registerFake(provider, f)
	cfg := fmt.Sprintf("MODE: butler\n"+
		"PERSON: \"e2e persona\"\n"+
		"SELECTED_PROVIDER: %s\n"+
		"PROVIDERS:\n"+
		"  %s:\n"+
		"    TYPE: anthropic\n"+
		"    MODEL: claude-opus-4-8\n"+
		"    URL: \"%s\"\n"+
		"    API_KEY: test-key\n",
		provider, provider, f.URL())
	return sc.writeFile("configs/butler.yaml", []byte(cfg))
}

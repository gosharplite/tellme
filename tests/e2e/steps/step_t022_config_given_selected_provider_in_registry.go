package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T022 — Given: a well-formed configuration "{config_path}" whose selected provider "{provider}" is in its registry
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a well-formed configuration "([^"]*)" whose selected provider "([^"]*)" is in its registry$`, givenWellFormedWithProvider)
	})
}

// givenWellFormedWithProvider writes a usable YAML whose SELECTED_PROVIDER is
// {provider} and whose PROVIDERS registry contains {provider} (怎麼做 /
// 權威狀態落地: the file's selected provider resolves to a registry entry).
func givenWellFormedWithProvider(ctx context.Context, configPath, provider string) error {
	return scenarioFrom(ctx).writeFile(configPath, []byte(wellFormedConfig("butler", provider)))
}

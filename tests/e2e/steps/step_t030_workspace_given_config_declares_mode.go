package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T030 — Given: the configuration "{config_path}" declares the mode "{mode}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the configuration "([^"]*)" declares the mode "([^"]*)"$`, givenConfigDeclaresMode)
	})
}

// givenConfigDeclaresMode writes a resolvable YAML whose MODE is {mode} and
// whose PROVIDERS registry contains the selected provider — so config load +
// validation succeed (怎麼做 / 權威狀態落地 / 回寫: file loadable and valid).
func givenConfigDeclaresMode(ctx context.Context, configPath, mode string) error {
	return scenarioFrom(ctx).writeFile(configPath, []byte(wellFormedConfig(mode, "deepseek-flash")))
}

package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T023 — Given: a well-formed configuration "{config_path}" whose provider registry is empty
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a well-formed configuration "([^"]*)" whose provider registry is empty$`, givenEmptyRegistryConfig)
	})
}

// givenEmptyRegistryConfig writes a usable YAML whose PROVIDERS registry is
// empty (怎麼做 / 權威狀態落地: the registry holds no entries).
func givenEmptyRegistryConfig(ctx context.Context, configPath string) error {
	return scenarioFrom(ctx).writeFile(configPath, []byte(emptyRegistryConfig("butler")))
}

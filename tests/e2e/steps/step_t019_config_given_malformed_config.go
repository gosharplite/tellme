package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T019 — Given: a malformed configuration "{config_path}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a malformed configuration "([^"]*)"$`, givenMalformedConfig)
	})
}

// givenMalformedConfig writes a file that exists but cannot be parsed as YAML
// (怎麼做 / 權威狀態落地 / 回寫: file exists but is unparseable).
func givenMalformedConfig(ctx context.Context, configPath string) error {
	return scenarioFrom(ctx).writeFile(configPath, []byte("MODE: butler\nPROVIDERS: [oops\n"))
}

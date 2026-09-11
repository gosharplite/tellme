package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T039 — Given: a configuration "{config_path}" that does not resolve
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configuration "([^"]*)" that does not resolve$`, givenUnresolvedConfig)
	})
}

// givenUnresolvedConfig arranges a configuration that does not reach a ready
// state (a malformed file → reason config-invalid) (怎麼做 / 權威狀態落地 /
// 回寫: setup resolution is not ready).
func givenUnresolvedConfig(ctx context.Context, configPath string) error {
	return scenarioFrom(ctx).writeFile(configPath, []byte("MODE: butler\nPROVIDERS: [oops\n"))
}

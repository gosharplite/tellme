package steps

import (
	"context"
	"github.com/cucumber/godog"
)

// T012 — Given: a well-formed configuration "{config_path}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a well-formed configuration "([^"]*)"$`, givenWellFormedConfig)
	})
}

// givenWellFormedConfig writes a usable YAML file (MODE, PERSON,
// SELECTED_PROVIDER, PROVIDERS) at the home-relative path; it parses and
// validates (怎麼做 / 權威狀態落地: the file exists and parses).
func givenWellFormedConfig(ctx context.Context, configPath string) error {
	sc := scenarioFrom(ctx)
	return sc.writeFile(configPath, []byte(wellFormedConfig("butler", "deepseek-flash")))
}

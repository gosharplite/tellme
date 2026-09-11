package steps

import (
	"context"
	"os"

	"github.com/cucumber/godog"
)

// T018 — Given: no configuration exists at "{config_path}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^no configuration exists at "([^"]*)"$`, givenNoConfigAt)
	})
}

// givenNoConfigAt ensures nothing exists at the home-relative path (怎麼做 /
// 權威狀態落地: the path is absent).
func givenNoConfigAt(ctx context.Context, configPath string) error {
	return os.RemoveAll(scenarioFrom(ctx).homePath(configPath))
}

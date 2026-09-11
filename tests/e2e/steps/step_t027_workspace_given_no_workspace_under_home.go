package steps

import (
	"context"
	"os"
	"path/filepath"

	"github.com/cucumber/godog"
)

// T027 — Given: no session workspace exists under "{home}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^no session workspace exists under "([^"]*)"$`, givenNoWorkspaceUnder)
	})
}

// givenNoWorkspaceUnder ensures the arranged runtime home has no output/
// directory (怎麼做 / 權威狀態落地: no session workspace exists).
func givenNoWorkspaceUnder(ctx context.Context, _ string) error {
	return os.RemoveAll(filepath.Join(scenarioFrom(ctx).home, "output"))
}

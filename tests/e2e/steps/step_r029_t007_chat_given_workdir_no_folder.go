package steps

import (
	"context"
	"os"
	"path/filepath"

	"github.com/cucumber/godog"
)

// T007 [BDD-RED] — Given: the working directory contains no folder "{name}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the working directory contains no folder "([^"]*)"$`, givenWorkdirNoFolder)
	})
}

// givenWorkdirNoFolder (怎麼做 / 權威狀態落地): ensure no folder named {name} exists
// in the subprocess working directory (remove it — and anything under it — if
// present).
func givenWorkdirNoFolder(ctx context.Context, name string) error {
	sc := scenarioFrom(ctx)
	return os.RemoveAll(filepath.Join(sc.workDir, filepath.FromSlash(name)))
}

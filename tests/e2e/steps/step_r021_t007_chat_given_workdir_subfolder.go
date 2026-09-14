package steps

import (
	"context"
	"os"
	"path/filepath"

	"github.com/cucumber/godog"
)

// T007 [BDD-RED] — Given: the working directory contains a sub-folder "{name}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the working directory contains a sub-folder "([^"]*)"$`, givenWorkdirSubfolder)
	})
}

// givenWorkdirSubfolder (怎麼做 / 權威狀態落地): create an empty sub-folder named
// {name} in the subprocess working directory (nested names such as "src/pkg" and
// ".git" are supported).
func givenWorkdirSubfolder(ctx context.Context, name string) error {
	sc := scenarioFrom(ctx)
	return os.MkdirAll(filepath.Join(sc.workDir, filepath.FromSlash(name)), 0o755)
}

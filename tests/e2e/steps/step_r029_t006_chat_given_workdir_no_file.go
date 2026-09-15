package steps

import (
	"context"
	"os"
	"path/filepath"

	"github.com/cucumber/godog"
)

// T006 [BDD-RED] — Given: the working directory contains no file "{name}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the working directory contains no file "([^"]*)"$`, givenWorkdirNoFile)
	})
}

// givenWorkdirNoFile (怎麼做 / 權威狀態落地): ensure no file named {name} exists in
// the subprocess working directory (remove it if present).
func givenWorkdirNoFile(ctx context.Context, name string) error {
	sc := scenarioFrom(ctx)
	if err := os.Remove(filepath.Join(sc.workDir, filepath.FromSlash(name))); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

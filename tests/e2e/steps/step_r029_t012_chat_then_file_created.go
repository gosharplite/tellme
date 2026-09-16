package steps

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cucumber/godog"
)

// T012 [BDD-RED] — Then: tellme created the file "{name}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme created the file "([^"]*)"$`, thenFileCreated)
	})
}

// thenFileCreated (必查 權威狀態): the file {name} exists on disk after the run.
func thenFileCreated(ctx context.Context, name string) error {
	sc := scenarioFrom(ctx)
	info, err := os.Stat(filepath.Join(sc.workDir, filepath.FromSlash(name)))
	if err != nil {
		return fmt.Errorf("the file %q was not created: %w", name, err)
	}
	if info.IsDir() {
		return fmt.Errorf("%q is a directory, not a file", name)
	}
	return nil
}

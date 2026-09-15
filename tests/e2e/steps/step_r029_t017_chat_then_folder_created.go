package steps

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cucumber/godog"
)

// T017 [BDD-RED] — Then: the folder "{name}" was created
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the folder "([^"]*)" was created$`, thenFolderCreated)
	})
}

// thenFolderCreated (必查 權威狀態): the folder {name} exists on disk after the run.
func thenFolderCreated(ctx context.Context, name string) error {
	sc := scenarioFrom(ctx)
	info, err := os.Stat(filepath.Join(sc.workDir, filepath.FromSlash(name)))
	if err != nil {
		return fmt.Errorf("the folder %q was not created: %w", name, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%q exists but is not a folder", name)
	}
	return nil
}

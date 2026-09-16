package steps

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cucumber/godog"
)

// T009 [BDD-RED] — Then: tellme creates no file "{path}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme creates no file "([^"]*)"$`, thenCreatesNoFile)
	})
}

// thenCreatesNoFile (必查 權威狀態): the file {path} does not exist on disk after
// the run — a run refused for a cut-off reply must not create the file (the
// truncated tool call must never be dispatched).
func thenCreatesNoFile(ctx context.Context, path string) error {
	sc := scenarioFrom(ctx)
	p := filepath.Join(sc.workDir, filepath.FromSlash(path))
	if _, err := os.Stat(p); err == nil {
		return fmt.Errorf("expected no file %q, but a file exists", path)
	} else if !os.IsNotExist(err) {
		return err
	}
	return nil
}

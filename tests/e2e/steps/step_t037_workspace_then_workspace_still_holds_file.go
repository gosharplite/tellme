package steps

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cucumber/godog"
)

// T037 — Then: the workspace "{workspace_path}" still holds the file "{file_name}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the workspace "([^"]*)" still holds the file "([^"]*)"$`, thenWorkspaceStillHoldsFile)
	})
}

// thenWorkspaceStillHoldsFile (必查 權威狀態): {file_name} still exists with its
// sentinel body; 不該發生: the reuse run must not delete or overwrite state.
func thenWorkspaceStillHoldsFile(ctx context.Context, workspacePath, fileName string) error {
	sc := scenarioFrom(ctx)
	p := filepath.Join(sc.workspacePath(workspacePath), fileName)
	b, err := os.ReadFile(p)
	if err != nil {
		return fmt.Errorf("preserved file %s missing after run: %w", p, err)
	}
	if len(b) == 0 {
		return fmt.Errorf("preserved file %s is empty after run", p)
	}
	return nil
}

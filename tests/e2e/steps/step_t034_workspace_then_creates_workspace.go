package steps

import (
	"context"
	"fmt"
	"os"

	"github.com/cucumber/godog"
)

// T034 — Then: tellme creates the session workspace "{workspace_path}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme creates the session workspace "([^"]*)"$`, thenCreatesWorkspace)
	})
}

// thenCreatesWorkspace (必查 權威狀態): the resolved workspace directory now
// exists and is a directory.
func thenCreatesWorkspace(ctx context.Context, workspacePath string) error {
	sc := scenarioFrom(ctx)
	p := sc.workspacePath(workspacePath)
	info, err := os.Stat(p)
	if err != nil {
		return fmt.Errorf("workspace %s was not created: %w", p, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("workspace %s is not a directory", p)
	}
	return nil
}

package steps

import (
	"context"
	"os"
	"path/filepath"

	"github.com/cucumber/godog"
)

// T033 — Given: the workspace path "{workspace_path}" already exists as a regular file
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the workspace path "([^"]*)" already exists as a regular file$`, givenWorkspaceIsRegularFile)
	})
}

// givenWorkspaceIsRegularFile creates a regular file (not a directory) at the
// resolved workspace path (怎麼做 / 權威狀態落地 / 回寫: the workspace path is a
// file).
func givenWorkspaceIsRegularFile(ctx context.Context, workspacePath string) error {
	sc := scenarioFrom(ctx)
	p := sc.workspacePath(workspacePath)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, []byte("not a directory"), 0o644)
}

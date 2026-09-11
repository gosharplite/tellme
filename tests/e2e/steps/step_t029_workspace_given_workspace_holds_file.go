package steps

import (
	"context"
	"os"
	"path/filepath"

	"github.com/cucumber/godog"
)

// T029 — Given: the workspace "{workspace_path}" already holds a file "{file_name}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the workspace "([^"]*)" already holds a file "([^"]*)"$`, givenWorkspaceHoldsFile)
	})
}

// givenWorkspaceHoldsFile writes a non-empty sentinel file into the resolved
// workspace directory (怎麼做 / 權威狀態落地 / 回寫: the workspace holds
// pre-existing content).
func givenWorkspaceHoldsFile(ctx context.Context, workspacePath, fileName string) error {
	sc := scenarioFrom(ctx)
	p := filepath.Join(sc.workspacePath(workspacePath), fileName)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, []byte("sentinel-body"), 0o644)
}

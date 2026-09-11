package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T046 — Then: tellme reports the session workspace resolved to "{workspace_path}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme reports the session workspace resolved to "([^"]*)"$`, thenReportsWorkspaceResolved)
	})
}

// thenReportsWorkspaceResolved (必查 呈現結果): stdout reports the session
// workspace resolved to the resolved path.
func thenReportsWorkspaceResolved(ctx context.Context, workspacePath string) error {
	sc := scenarioFrom(ctx)
	p := sc.workspacePath(workspacePath)
	if !strings.Contains(sc.stdout, p) {
		return fmt.Errorf("stdout %q does not report the session workspace %q", sc.stdout, p)
	}
	return nil
}

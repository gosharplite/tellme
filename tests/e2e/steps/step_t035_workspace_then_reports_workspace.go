package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T035 — Then: tellme reports the session workspace "{workspace_path}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme reports the session workspace "([^"]*)"$`, thenReportsWorkspace)
	})
}

// thenReportsWorkspace (必查 呈現結果): stdout reports the resolved workspace
// path; 權威狀態: the reported path matches the directory that exists.
func thenReportsWorkspace(ctx context.Context, workspacePath string) error {
	sc := scenarioFrom(ctx)
	p := sc.workspacePath(workspacePath)
	if !strings.Contains(sc.stdout, p) {
		return fmt.Errorf("stdout %q does not report the session workspace %q", sc.stdout, p)
	}
	return nil
}

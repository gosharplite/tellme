package steps

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cucumber/godog"
)

// T036 — Then: tellme reuses the session workspace "{workspace_path}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme reuses the session workspace "([^"]*)"$`, thenReusesWorkspace)
	})
}

// thenReusesWorkspace (必查 權威狀態): the workspace directory is the same as
// before (same inode — not re-created); 呈現結果: the run reported the reused
// workspace.
func thenReusesWorkspace(ctx context.Context, workspacePath string) error {
	sc := scenarioFrom(ctx)
	p := sc.workspacePath(workspacePath)
	info, err := os.Stat(p)
	if err != nil {
		return fmt.Errorf("workspace %s missing after run: %w", p, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("workspace %s is not a directory", p)
	}
	if sc.wsSet {
		ino, err := inodeOf(p)
		if err != nil {
			return err
		}
		if ino != sc.wsIno {
			return fmt.Errorf("workspace %s was re-created (inode %d -> %d)", p, sc.wsIno, ino)
		}
	}
	if !strings.Contains(sc.stdout, p) {
		return fmt.Errorf("stdout %q does not report the reused workspace %q", sc.stdout, p)
	}
	return nil
}

package steps

import (
	"context"
	"os"

	"github.com/cucumber/godog"
)

// T028 — Given: the session workspace "{workspace_path}" already exists
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the session workspace "([^"]*)" already exists$`, givenWorkspaceExists)
	})
}

// givenWorkspaceExists resolves {workspace_path} and creates the directory
// (怎麼做 / 權威狀態落地 / 回寫: the workspace directory exists). Its inode is
// recorded so a reuse assertion can prove it was not re-created.
func givenWorkspaceExists(ctx context.Context, workspacePath string) error {
	sc := scenarioFrom(ctx)
	p := sc.workspacePath(workspacePath)
	if err := os.MkdirAll(p, 0o755); err != nil {
		return err
	}
	if ino, err := inodeOf(p); err == nil {
		sc.wsIno = ino
		sc.wsSet = true
	}
	return nil
}

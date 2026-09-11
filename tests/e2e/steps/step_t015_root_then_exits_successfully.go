package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/internal/cli"
)

// T015 — Then: tellme exits successfully
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme exits successfully$`, thenExitsSuccessfully)
	})
}

// thenExitsSuccessfully (必查 呈現結果): the exit code equals the success code.
func thenExitsSuccessfully(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if sc.runErr != nil {
		return fmt.Errorf("tellme run error: %w", sc.runErr)
	}
	if sc.exitCode != cli.Success {
		return fmt.Errorf("exit code = %d, want %d (success); stderr=%q", sc.exitCode, cli.Success, sc.stderr)
	}
	return nil
}

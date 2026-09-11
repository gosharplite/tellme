package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/internal/cli"
)

// T016 — Then: tellme refuses to proceed
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme refuses to proceed$`, thenRefusesToProceed)
	})
}

// thenRefusesToProceed (必查 呈現結果): the exit code is non-zero.
func thenRefusesToProceed(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if sc.runErr != nil {
		return fmt.Errorf("tellme run error: %w", sc.runErr)
	}
	if sc.exitCode == cli.Success {
		return fmt.Errorf("exit code = 0, want non-zero (refused); stdout=%q", sc.stdout)
	}
	return nil
}

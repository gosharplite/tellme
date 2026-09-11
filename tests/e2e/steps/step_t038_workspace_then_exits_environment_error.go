package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/internal/cli"
)

// T038 — Then: tellme exits with the environment error code
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme exits with the environment error code$`, thenExitsEnvironmentError)
	})
}

// thenExitsEnvironmentError (必查 呈現結果): the exit code equals its dedicated
// environment-error code and does not collapse to success or another class.
func thenExitsEnvironmentError(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if sc.exitCode != cli.EnvironmentError {
		return fmt.Errorf("exit code = %d, want %d (environment error)", sc.exitCode, cli.EnvironmentError)
	}
	return nil
}

package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/internal/cli"
)

// T054 — Then: tellme exits with the usage error code
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme exits with the usage error code$`, thenExitsUsageError)
	})
}

// thenExitsUsageError (必查 呈現結果): the exit code equals its dedicated usage
// error code and does not collapse to success or another class.
func thenExitsUsageError(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if sc.exitCode != cli.UsageError {
		return fmt.Errorf("exit code = %d, want %d (usage error)", sc.exitCode, cli.UsageError)
	}
	return nil
}

package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/internal/cli"
)

// T026 — Then: tellme exits with the configuration error code
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme exits with the configuration error code$`, thenExitsConfigError)
	})
}

// thenExitsConfigError (必查 呈現結果): the exit code equals its dedicated
// configuration-error code and does not collapse to success or another class.
func thenExitsConfigError(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if sc.exitCode != cli.ConfigError {
		return fmt.Errorf("exit code = %d, want %d (configuration error)", sc.exitCode, cli.ConfigError)
	}
	return nil
}

package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/internal/cli"
)

// T019 — Then: tellme exits with the provider error code
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme exits with the provider error code$`, thenExitsProviderError)
	})
}

// thenExitsProviderError (必查 呈現結果): the exit code equals the pinned
// provider error code (6) and does not collapse to success or another class.
func thenExitsProviderError(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if sc.exitCode != cli.ProviderError {
		return fmt.Errorf("exit code = %d, want %d (provider error)", sc.exitCode, cli.ProviderError)
	}
	return nil
}

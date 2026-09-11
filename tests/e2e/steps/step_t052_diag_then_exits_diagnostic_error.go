package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/internal/cli"
)

// T052 — Then: tellme exits with the diagnostic error code
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme exits with the diagnostic error code$`, thenExitsDiagnosticError)
	})
}

// thenExitsDiagnosticError (必查 呈現結果): the report was produced and the exit
// code equals its dedicated diagnostic "unresolved" code.
func thenExitsDiagnosticError(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if sc.exitCode != cli.DiagnosticUnresolvedError {
		return fmt.Errorf("exit code = %d, want %d (diagnostic: unresolved)", sc.exitCode, cli.DiagnosticUnresolvedError)
	}
	return nil
}

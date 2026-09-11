package steps

import (
	"context"
	"fmt"
	"os"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/harness"
)

// T010 — Given: the operator has a runnable tellme installation
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the operator has a runnable tellme installation$`, givenRunnableInstallation)
	})
}

// givenRunnableInstallation asserts the once-built tellme binary exists and is
// executable (怎麼做 / 權威狀態落地: a runnable binary is available for the scenario).
func givenRunnableInstallation(context.Context) error {
	bin, err := harness.BinaryPath()
	if err != nil {
		return fmt.Errorf("build tellme binary: %w", err)
	}
	info, err := os.Stat(bin)
	if err != nil {
		return fmt.Errorf("stat tellme binary %s: %w", bin, err)
	}
	if info.Mode()&0o111 == 0 {
		return fmt.Errorf("tellme binary %s is not executable", bin)
	}
	return nil
}

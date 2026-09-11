package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T051 — Then: tellme performs no network access
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme performs no network access$`, thenPerformsNoNetworkAccess)
	})
}

// thenPerformsNoNetworkAccess is a harness-mediated witness, not a black-box
// observation (必查 呈現結果 / 權威狀態): the build graph links no
// network-capable package, AND rerunning under a hostile network environment
// leaves the exit code and stdout unchanged.
func thenPerformsNoNetworkAccess(ctx context.Context) error {
	sc := scenarioFrom(ctx)

	if pkg, err := networkCapabilityViolation(); err != nil {
		return fmt.Errorf("build-graph guard: %w", err)
	} else if pkg != "" {
		return fmt.Errorf("network-capable package %q is linked into ./cmd/tellme", pkg)
	}

	blocked := sc.blockedRun()
	if blocked.Err != nil {
		return fmt.Errorf("blocked-network run error: %w", blocked.Err)
	}
	if blocked.ExitCode != sc.exitCode || blocked.Stdout != sc.stdout {
		return fmt.Errorf("hostile network changed the outcome: normal(exit=%d) blocked(exit=%d)", sc.exitCode, blocked.ExitCode)
	}
	return nil
}

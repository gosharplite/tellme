package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T051 — Then: tellme performs no network access (round-004 re-scope)
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme performs no network access$`, thenPerformsNoNetworkAccess)
	})
}

// thenPerformsNoNetworkAccess is a harness-mediated offline-path witness
// (必查 呈現結果 / 權威狀態), not a black-box observation: (i) any fake provider
// registered for the scenario — the recording sink the configured endpoint
// points at — must have received ZERO requests, and (ii) rerunning under a
// hostile network environment leaves the exit code and stdout unchanged.
//
// Round 004 retired the whole-binary capability guard (the prompt-bearing chat
// path legitimately links net/http), so this is an offline-path behaviour claim
// rather than a whole-binary absence claim.
func thenPerformsNoNetworkAccess(ctx context.Context) error {
	sc := scenarioFrom(ctx)

	for name, f := range sc.fakeByProvider {
		if n := f.RequestCount(); n != 0 {
			return fmt.Errorf("offline path contacted provider %q (%d requests)", name, n)
		}
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

package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T003 [BDD-RED] — Given: the diagnostics are shown at a terminal (interface-root row).
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the diagnostics are shown at a terminal$`, givenDiagnosticsAtTerminal)
	})
}

// givenDiagnosticsAtTerminal forces tellme's diagnostic-stream (stderr) terminal
// probe to report a terminal via the TELL_ME_FORCE_STDERR_TTY diagnostic seam
// (round-019 interface-root row), so the spinner is drivable end-to-end without
// a pty (怎麼做: the seam is set for the run; 權威狀態落地: the effective stderr
// probe reports a terminal; 回寫: none).
func givenDiagnosticsAtTerminal(ctx context.Context) error {
	scenarioFrom(ctx).setEnv("TELL_ME_FORCE_STDERR_TTY", "1")
	return nil
}

package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T004 [BDD-RED] (round 025) — Given: the diagnostics are shown at a terminal
// narrower than the indicator line.
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the diagnostics are shown at a terminal narrower than the indicator line$`, givenNarrowTerminal)
	})
}

// givenNarrowTerminal (怎麼做 / 權威狀態落地 / 回寫): force tellme's diagnostic-stream
// (stderr) terminal probe to report a terminal AND its width probe to report a
// narrow width, so a tool-phase spinner frame soft-wraps across rows. Sets
// TELL_ME_FORCE_STDERR_TTY=1 and TELL_ME_FORCE_STDERR_COLS to a narrow width.
func givenNarrowTerminal(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	sc.setEnv("TELL_ME_FORCE_STDERR_TTY", "1")
	sc.setEnv("TELL_ME_FORCE_STDERR_COLS", "40")
	return nil
}

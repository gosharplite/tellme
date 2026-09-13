package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// Round-012 review RF1 — Given: the operator is working at an interactive terminal.
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the operator is working at an interactive terminal$`, givenInteractiveTerminal)
	})
}

// givenInteractiveTerminal forces tellme's terminal probe to report the stream as
// a terminal via the TELL_ME_FORCE_STDIN_TTY diagnostic seam, so the interactive
// multi-line read can be exercised end-to-end over a pipe without a pty
// (怎麼做: the seam is set for the run; 回寫: none).
func givenInteractiveTerminal(ctx context.Context) error {
	scenarioFrom(ctx).setEnv("TELL_ME_FORCE_STDIN_TTY", "1")
	return nil
}

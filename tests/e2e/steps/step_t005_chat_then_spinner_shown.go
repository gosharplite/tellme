package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T005 [BDD-RED] — Then: the run shows the progress spinner while it waits.
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the run shows the progress spinner while it waits$`, thenSpinnerShown)
	})
}

// thenSpinnerShown (必查 / 呈現結果): the captured stderr carries a spinner line — a
// braille frame followed by a phase status and the round-040 DUAL `(<total>s <call>s)`
// elapsed segment (the carriage-return redraw leaves the latest frame; a fast turn
// may render a single synchronous frame). 不該發生: the spinner must not be written
// to stdout.
func thenSpinnerShown(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if !reSpinnerLine.MatchString(sc.stderr) {
		return fmt.Errorf("the diagnostic stream carries no progress spinner: stderr=%q", sc.stderr)
	}
	if hasSpinnerFrame(sc.stdout) {
		return fmt.Errorf("the progress spinner was written to standard output: stdout=%q", sc.stdout)
	}
	return nil
}

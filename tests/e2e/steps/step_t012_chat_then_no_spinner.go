package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T012 [BDD-RED] — Then: the run shows no progress spinner (interface-root row).
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the run shows no progress spinner$`, thenRunNoProgressSpinner)
	})
}

// thenRunNoProgressSpinner (必查 / 呈現結果): neither the captured standard output
// nor the standard error shows a progress-spinner residue (a braille frame
// followed by a phase status). Used for the non-terminal `stderr`, `-r`,
// `--version`, `-l`, prompt-less `--new`, and the `-i` submit path — and the
// failed-turn carrier, where the failure is detected mid-wait. 不該發生: the
// spinner must not survive (residue) on those surfaces.
//
// The stderr check simulates the carriage-return redraw (spinnerVisible): a
// synchronously-drawn frame that the run cleared leaves no visible residue, so
// the failed-turn carrier passes without weakening the other negatives (which
// never draw at all). A spinner that is drawn and NOT cleared remains visible and
// fails this step (the falsifiability witness).
func thenRunNoProgressSpinner(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if hasSpinnerFrame(sc.stdout) {
		return fmt.Errorf("the run shows a progress spinner on standard output: %q", sc.stdout)
	}
	if hasSpinnerFrame(spinnerVisible(sc.stderr)) {
		return fmt.Errorf("the run shows a progress spinner residue on the diagnostic stream: %q", sc.stderr)
	}
	return nil
}

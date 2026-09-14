package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T008 — Then: the interactive prompt marks one suggestion as the current choice
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the interactive prompt marks one suggestion as the current choice$`, thenMarksCurrentChoice)
	})
}

// thenMarksCurrentChoice (必查 呈現結果): the captured output renders the `>`
// selection cursor on a suggestion row. The capture accumulates rendered frames,
// so per-frame exactness is the unit pin (T017); this asserts presence.
func thenMarksCurrentChoice(ctx context.Context) error {
	if out := renderedOutput(scenarioFrom(ctx)); tuiCursorRows(out) < 1 {
		return fmt.Errorf("the interactive prompt did not mark a suggestion as the current choice; output=%q", out)
	}
	return nil
}

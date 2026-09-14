package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T012 — Then: the interactive prompt frames the editor within the terminal width
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the interactive prompt frames the editor within the terminal width$`, thenFramesWithinWidth)
	})
}

// thenFramesWithinWidth (必查 呈現結果): the captured output draws the editor
// frame and no visible line exceeds a bounded terminal width (the editor is
// drawn to the run's width, not unbounded).
func thenFramesWithinWidth(ctx context.Context) error {
	out := renderedOutput(scenarioFrom(ctx))
	if !tuiHasBorder(out) {
		return fmt.Errorf("the interactive prompt did not draw the editor frame; output=%q", out)
	}
	if w := tuiMaxLineWidth(out); w > 120 {
		return fmt.Errorf("the editor frame exceeded the terminal width (%d columns); output=%q", w, out)
	}
	return nil
}

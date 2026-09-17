package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T008 — Then: the interactive prompt marks no suggestion as the current choice
// (round 037; supersedes the round-016 "marks one suggestion as the current choice").
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the interactive prompt marks no suggestion as the current choice$`, thenMarksNoCurrentChoice)
	})
}

// thenMarksNoCurrentChoice (必查 呈現結果): the rendered prompt's suggestion block carries
// NO `>` cursor row at rest (round 037: the selection starts at a no-choice sentinel and
// resets on every refresh). The carrier is the precise cursor-row predicate
// (`tuiCursorRows`, an exact `"  > "`-prefix match — so a suggestion whose text begins
// with `> ` is not miscounted; round-037 review G-2) over the captured output. The
// after-`Tab` selection (a legitimate cursor row) is carried by the unit pins, not here;
// per-frame exactness remains the unit pin too.
func thenMarksNoCurrentChoice(ctx context.Context) error {
	if n := tuiCursorRows(renderedOutput(scenarioFrom(ctx))); n != 0 {
		return fmt.Errorf("the rendered suggestion block marked %d suggestion(s) as the current choice, want 0", n)
	}
	return nil
}

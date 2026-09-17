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
// no `>` cursor row (round 037: the selection starts at a no-choice sentinel and resets on
// every refresh). The predicate is the exact rendered cursor row (`tuiCursorRows`, a
// `"  > "`-prefix match — a suggestion whose text begins with `> ` is not miscounted;
// round-037 review G-2).
//
// PRECONDITION: the scenario must not navigate. Round-037 review G-1 established that the
// harness delivers the compose keys (incl. any `Tab`) before the first paint, so a
// navigating scenario carries one painted frame and there is no pre-navigation frame to
// scope to; the after-`Tab` selection is carried by the unit pins (TestTabFromNoChoiceSelectsFirst,
// TestCycleArithmetic). Per-frame exactness is likewise a unit pin.
func thenMarksNoCurrentChoice(ctx context.Context) error {
	if n := tuiCursorRows(renderedOutput(scenarioFrom(ctx))); n != 0 {
		return fmt.Errorf("the rendered suggestion block marked %d suggestion(s) as the current choice, want 0", n)
	}
	return nil
}

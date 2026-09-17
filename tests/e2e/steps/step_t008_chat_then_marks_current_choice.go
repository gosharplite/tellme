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

// thenMarksNoCurrentChoice (必查 呈現結果): the AT-REST frame — the first painted frame,
// before any key is delivered — renders NO `>` selection cursor on any suggestion row
// (round 037: the selection starts at a no-choice sentinel and resets on every refresh).
// The scope is the at-rest frame, not the whole accumulated stream, so a legitimate
// post-navigation cursor (a later frame after a `Tab`) cannot false-fail this rule
// (round-037 review F-1); per-frame exactness remains the unit pin.
func thenMarksNoCurrentChoice(ctx context.Context) error {
	if n := tuiCursorRows(atRestFrame(renderedOutput(scenarioFrom(ctx)))); n != 0 {
		return fmt.Errorf("the at-rest prompt marked %d suggestion(s) as the current choice, want 0", n)
	}
	return nil
}

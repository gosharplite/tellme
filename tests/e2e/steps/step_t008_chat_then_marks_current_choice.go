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

// thenMarksNoCurrentChoice (必查 呈現結果): at rest the captured output renders NO
// `>` selection cursor on any suggestion row (round 037 — the selection starts at a
// no-choice sentinel and resets on every refresh). The capture accumulates rendered
// frames, so per-frame exactness is the unit pin; this asserts the whole stream.
func thenMarksNoCurrentChoice(ctx context.Context) error {
	if n := tuiCursorRows(renderedOutput(scenarioFrom(ctx))); n != 0 {
		return fmt.Errorf("the interactive prompt marked %d suggestion(s) as the current choice at rest, want 0", n)
	}
	return nil
}

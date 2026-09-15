package steps

import (
	"context"
	"fmt"
	"regexp"

	"github.com/cucumber/godog"
)

// T005 [BDD-RED] (round 025) — Then: the progress indicator is cleared from every
// row it occupied.
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the progress indicator is cleared from every row it occupied$`, thenClearedFromEveryRow)
	})
}

// reCursorUp matches an ANSI cursor-up control (`\x1b[<n>A`, n optional), emitted
// ONLY by the round-025 row-aware clear.
var reCursorUp = regexp.MustCompile("\x1b\\[[0-9]*A")

// thenClearedFromEveryRow (必查 / 呈現結果): the captured stderr carries a clear
// sequence that erases EVERY terminal row the last frame occupied — with the
// narrow-width seam forcing a frame wider than the terminal, the clear includes a
// cursor-up so no wrapped row survives. 不該發生: a single-row clear (no cursor-up)
// must not be used for a wrapped frame (residue).
func thenClearedFromEveryRow(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if !reCursorUp.MatchString(sc.stderr) {
		return fmt.Errorf("the clear erased a single row only (no cursor-up) for a wrapped frame: stderr=%q", sc.stderr)
	}
	return nil
}

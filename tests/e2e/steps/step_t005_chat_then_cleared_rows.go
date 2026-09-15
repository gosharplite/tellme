package steps

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/cucumber/godog"
)

// T005 [BDD-RED] (round 025) — Then: the progress indicator is cleared from every
// row it occupied.
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the progress indicator is cleared from every row it occupied$`, thenClearedFromEveryRow)
	})
}

// reToolLogStart matches the start of a tool-loop log line (`[HH:MM:SS] [Tool] `).
var reToolLogStart = regexp.MustCompile(`\[\d{2}:\d{2}:\d{2}\] \[Tool\] `)

// thenClearedFromEveryRow (必查 / 呈現結果): the erase immediately preceding a
// tool-loop log line — the yield-to-output clear of the (narrow-width-forced)
// WRAPPED tool-phase frame — is rows-aware: within the bytes just before the log
// line there is a cursor-up (`\x1b[1A`), so it erased the wrapped row above as
// well as the bottom row. 不該發生: a single-row clear (no cursor-up) must not be
// used to yield a wrapped frame (residue).
//
// Correction (round-025 review TD-1): the erase immediately preceding the log
// line is emitted by clearLocked, so a clearLocked-only single-row regression —
// which a whole-turn cursor-up presence check (the Redraw path also emits
// cursor-ups) or a frame-boundary check (renderLocked also emits them) would
// miss — is caught here. The exact `rows = ceil(width/columns)` sequence remains
// pinned by the internal/ui unit test (TestSpinnerRowAwareClear); a flat byte
// capture has no terminal grid, so it cannot witness residue directly.
func thenClearedFromEveryRow(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	loc := reToolLogStart.FindStringIndex(sc.stderr)
	if loc == nil {
		return fmt.Errorf("the run reported no tool-loop log line to bound the clear against: stderr=%q", sc.stderr)
	}
	pre := sc.stderr[:loc[0]]
	if len(pre) > 16 {
		pre = pre[len(pre)-16:]
	}
	if !strings.Contains(pre, "\x1b[1A") {
		return fmt.Errorf("the erase before the tool-loop log line was single-row (no cursor-up): stderr=%q", sc.stderr)
	}
	return nil
}

package steps

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cucumber/godog"
)

// T018 [BDD-RED] — Then: the run reported a tool step marker for each of the {count} executed rounds
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the run reported a tool step marker for each of the ([0-9]+) executed rounds$`, thenStepMarkerPerRound)
	})
}

// thenStepMarkerPerRound (必查 呈現結果): the count of `[Tool Engine] Step i/M`
// lines equals the number of executed rounds — `i` and `M` are executed-round
// units (round 034 FR-001 / ADR 0005 D9 / G9).
func thenStepMarkerPerRound(ctx context.Context, count string) error {
	sc := scenarioFrom(ctx)
	want, err := strconv.Atoi(count)
	if err != nil {
		return fmt.Errorf("invalid executed-round count %q: %w", count, err)
	}
	got := toolEngineLineCount(sc.stderr)
	if got != want {
		return fmt.Errorf("expected %d `[Tool Engine]` step markers, got %d; stderr=%q", want, got, sc.stderr)
	}
	return nil
}

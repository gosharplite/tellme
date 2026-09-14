package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T009 — Then: the run shows no turn chrome
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the run shows no turn chrome$`, thenNoTurnChrome)
	})
}

// thenNoTurnChrome (必查 呈現結果): neither the captured standard output nor
// standard error carries the input-capture acknowledgement, the 80-column rule,
// or a `╭─⠿ Turn …` header — the `-i` prompt and the non-prompt paths show none.
func thenNoTurnChrome(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if out := renderedOutput(sc); hasAnyTurnChrome(out) {
		return fmt.Errorf("the run rendered turn chrome; output=%q", out)
	}
	return nil
}

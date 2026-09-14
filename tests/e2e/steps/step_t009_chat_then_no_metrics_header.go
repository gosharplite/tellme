package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T009 — Then: the interactive prompt shows no session metrics header
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the interactive prompt shows no session metrics header$`, thenNoSessionMetricsHeader)
	})
}

// thenNoSessionMetricsHeader (必查 呈現結果): the captured output carries NO line
// matching the specific dashboard pattern `provider … | tokens … | turns …`
// (round-016 F3: a specific pattern, not a bare substring). The non-vacuity
// witness (T026(a)) re-adds the header → this Then must fail.
func thenNoSessionMetricsHeader(ctx context.Context) error {
	if out := renderedOutput(scenarioFrom(ctx)); tuiHasMetricsHeader(out) {
		return fmt.Errorf("the interactive prompt rendered a session metrics header; output=%q", out)
	}
	return nil
}

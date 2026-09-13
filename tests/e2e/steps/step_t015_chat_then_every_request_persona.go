package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T015 — Then: every request carried the persona "{persona}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^every request carried the persona "([^"]*)"$`, thenEveryRequestCarriedPersona)
	})
}

// thenEveryRequestCarriedPersona (必查 呈現結果): every request the fake recorded
// leads with a `system`-role message whose content equals {persona} — the
// round-011 FR-004 assertion that the persona rides every request of the turn,
// including tool-driven completions.
func thenEveryRequestCarriedPersona(ctx context.Context, persona string) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	n := f.RequestCount()
	if n == 0 {
		return fmt.Errorf("no request was recorded")
	}
	for i := 0; i < n; i++ {
		msgs, err := decodeWireMessages(f.BodyAt(i))
		if err != nil {
			return fmt.Errorf("decode request %d: %w", i, err)
		}
		if len(msgs) == 0 || msgs[0].Role != "system" || msgs[0].Content != persona {
			return fmt.Errorf("request %d did not lead with the persona %q; messages=%+v", i, persona, msgs)
		}
	}
	return nil
}

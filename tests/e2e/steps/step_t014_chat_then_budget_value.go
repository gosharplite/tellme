package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T014 — Then: the payload status measures against a budget of {budget} tokens
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the payload status measures against a budget of ([0-9]+) tokens$`, thenPayloadBudgetValue)
	})
}

// thenPayloadBudgetValue (必查 呈現結果): the `<max>` field of the reported
// payload status line equals {budget}.
func thenPayloadBudgetValue(ctx context.Context, budget int) error {
	sc := scenarioFrom(ctx)
	got, ok := payloadStatusBudget(sc.stderr)
	if !ok {
		return fmt.Errorf("standard error carried no payload status line; stderr=%q", sc.stderr)
	}
	if got != budget {
		return fmt.Errorf("payload status budget = %d, want %d", got, budget)
	}
	return nil
}

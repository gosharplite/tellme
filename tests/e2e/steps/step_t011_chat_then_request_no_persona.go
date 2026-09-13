package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T011 — Then: the request carried no persona
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the request carried no persona$`, thenRequestCarriedNoPersona)
	})
}

// thenRequestCarriedNoPersona (必查 呈現結果): the recorded request's messages carry
// no `system`-role message.
func thenRequestCarriedNoPersona(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	msgs, err := decodeWireMessages(f.LastBody())
	if err != nil {
		return fmt.Errorf("decode request messages: %w", err)
	}
	for _, m := range msgs {
		if m.Role == "system" {
			return fmt.Errorf("the request carried a system message, want none: %+v", m)
		}
	}
	return nil
}

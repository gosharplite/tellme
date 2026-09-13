package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T010 — Then: the request carried the persona "{persona}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the request carried the persona "([^"]*)"$`, thenRequestCarriedPersona)
	})
}

// thenRequestCarriedPersona (必查 呈現結果): the recorded request's messages begin
// with a `system`-role message whose content equals {persona}.
func thenRequestCarriedPersona(ctx context.Context, persona string) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	msgs, err := decodeWireMessages(f.LastBody())
	if err != nil {
		return fmt.Errorf("decode request messages: %w", err)
	}
	if len(msgs) == 0 || msgs[0].Role != "system" || msgs[0].Content != persona {
		return fmt.Errorf("the request did not lead with the persona %q; messages=%+v", persona, msgs)
	}
	return nil
}

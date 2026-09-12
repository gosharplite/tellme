package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T010 — Then: the request carried no earlier exchange
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the request carried no earlier exchange$`, thenRequestCarriedNoEarlierExchange)
	})
}

// thenRequestCarriedNoEarlierExchange (必查 呈現結果): the fake recorded exactly one
// request whose messages are exactly the current user prompt.
func thenRequestCarriedNoEarlierExchange(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	msgs, err := decodeWireMessages(f.LastBody())
	if err != nil {
		return fmt.Errorf("decode request messages: %w", err)
	}
	if len(msgs) != 1 {
		return fmt.Errorf("the request carried %d messages, want exactly the current prompt; %+v", len(msgs), msgs)
	}
	if msgs[0].Role != "user" || msgs[0].Content != sc.lastPrompt {
		return fmt.Errorf("the single message = %+v, want the current user prompt %q", msgs[0], sc.lastPrompt)
	}
	return nil
}

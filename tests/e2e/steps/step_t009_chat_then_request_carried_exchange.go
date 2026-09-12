package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T009 — Then: the request carried the earlier exchange "{prompt}" and "{answer}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the request carried the earlier exchange "([^"]*)" and "([^"]*)"$`, thenRequestCarriedEarlierExchange)
	})
}

// thenRequestCarriedEarlierExchange (必查 呈現結果 / 權威狀態): the fake recorded a
// request whose messages carry the earlier user message {prompt} and the earlier
// assistant message {answer}, in order, ahead of the current prompt.
func thenRequestCarriedEarlierExchange(ctx context.Context, prompt, answer string) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	msgs, err := decodeWireMessages(f.LastBody())
	if err != nil {
		return fmt.Errorf("decode request messages: %w", err)
	}
	for i := 0; i+1 < len(msgs); i++ {
		if msgs[i].Role == "user" && msgs[i].Content == prompt &&
			msgs[i+1].Role == "assistant" && msgs[i+1].Content == answer {
			return nil
		}
	}
	return fmt.Errorf("the request did not carry the earlier exchange %q/%q; messages=%+v", prompt, answer, msgs)
}

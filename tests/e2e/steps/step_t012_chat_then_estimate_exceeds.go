package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/internal/domain/llm"
)

// T012 — Then: the estimated payload exceeds the conversation messages alone
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the estimated payload exceeds the conversation messages alone$`, thenEstimateExceedsMessages)
	})
}

// thenEstimateExceedsMessages (必查 呈現結果): the reported pre-flight estimate
// `~<n>` is strictly greater than the estimator applied to the recorded
// request's messages excluding any leading `system` message — i.e. it also
// counts the persona and the offered tools.
func thenEstimateExceedsMessages(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	n, ok := estimatedPayloadValue(sc.stderr)
	if !ok {
		return fmt.Errorf("standard error carried no estimated payload status line; stderr=%q", sc.stderr)
	}
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	msgs, err := decodeWireMessages(f.LastBody())
	if err != nil {
		return fmt.Errorf("decode request messages: %w", err)
	}
	convo := make([]llm.Message, 0, len(msgs))
	for _, m := range msgs {
		if m.Role == "system" {
			continue
		}
		convo = append(convo, llm.Message{Role: m.Role, Content: m.Content})
	}
	if base := llm.EstimateTokens(convo); n <= base {
		return fmt.Errorf("the estimated payload %d must exceed the conversation-messages-only estimate %d", n, base)
	}
	return nil
}

package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T009 [BDD-RED] — Then: the run asked for a reason before running a tool
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the run asked for a reason before running a tool$`, thenRunAskedForAReason)
	})
}

// thenRunAskedForAReason (必查 呈現結果): a reason-less tool call was REFUSED —
// the loop folded back a recoverable result asking the model to retry with a
// reason (round 056 / ADR 0025 D3). The recoverable result is observable in the
// recorded request conversation as a `tool`-role message carrying the ask.
func thenRunAskedForAReason(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no provider fake recorded a request")
	}
	const marker = "a reason is required to call a tool"
	for i := 0; i < f.RequestCount(); i++ {
		for _, msg := range f.MessagesAt(i) {
			role, _ := msg["role"].(string)
			if role != "tool" {
				continue
			}
			content, _ := msg["content"].(string)
			if strings.Contains(content, marker) {
				return nil
			}
		}
	}
	return fmt.Errorf("no recoverable `%s` result was fed back to the model", marker)
}

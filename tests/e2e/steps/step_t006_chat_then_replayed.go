package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T006 — Then: the request replayed the earlier tool step "{tool}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the request replayed the earlier tool step "([^"]*)"$`, thenReplayedToolStep)
	})
}

// thenReplayedToolStep (必查 呈現結果 / 權威狀態): a recorded request replayed the
// earlier tool step {tool} — an assistant tool call for {tool} followed by its
// tool result — wire-family-agnostic (reuses the round-008 tool-message helpers).
func thenReplayedToolStep(ctx context.Context, tool string) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	if !hasToolCall(f, tool, "") {
		return fmt.Errorf("the request did not replay the earlier tool step %q; body=%s", tool, f.LastBody())
	}
	return nil
}

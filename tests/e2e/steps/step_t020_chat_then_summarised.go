package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T020 — Then: tellme summarised the earlier conversation using its summarise tool
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme summarised the earlier conversation using its summarise tool$`, thenSummarised)
	})
}

// thenSummarised (必查 權威狀態): the fake recorded the model's `summarize_history`
// tool call and the run fed the produced summary back into the conversation (a
// non-error tool result). Summarisation must be an on-demand tool.
func thenSummarised(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	if !hasToolCall(f, "summarize_history", "") {
		return fmt.Errorf("the run did not make a summarize_history tool call")
	}
	result := lastToolResult(f)
	if result == "" || strings.HasPrefix(result, "error:") {
		return fmt.Errorf("the summarise tool did not return a summary; result=%q", result)
	}
	return nil
}

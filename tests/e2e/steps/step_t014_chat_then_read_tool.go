package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T014 — Then: tellme read "{path}" using its read_files tool
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme read "([^"]*)" using its read_files tool$`, thenReadTool)
	})
}

// thenReadTool (必查 權威狀態): the fake recorded the model's `read_files` tool call
// for {path} and the run fed that tool's result back into the conversation (a
// later request carried a tool result).
func thenReadTool(ctx context.Context, path string) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	if !hasToolCall(f, "read_files", path) {
		return fmt.Errorf("the run did not make a read_files tool call for %q", path)
	}
	if countToolIterations(f) == 0 {
		return fmt.Errorf("the read_files result was not fed back into the conversation")
	}
	return nil
}

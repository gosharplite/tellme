package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T021 [BDD-RED] — Then: tellme listed the directory using its list_files tool
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme listed the directory using its list_files tool$`, thenListed)
	})
}

// thenListed (必查 權威狀態): the fake recorded the model's list_files tool call and
// the run fed its result back into the conversation.
func thenListed(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	if !hasToolCall(f, "list_files", "") {
		return fmt.Errorf("the run did not make a list_files tool call")
	}
	if countToolIterations(f) == 0 {
		return fmt.Errorf("the list_files result was not fed back into the conversation")
	}
	return nil
}

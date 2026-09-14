package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T024 [BDD-RED] — Then: tellme showed the folder tree using its get_tree tool
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme showed the folder tree using its get_tree tool$`, thenTree)
	})
}

// thenTree (必查 權威狀態): the fake recorded the model's get_tree tool call and the
// run fed its result back into the conversation.
func thenTree(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	if !hasToolCall(f, "get_tree", "") {
		return fmt.Errorf("the run did not make a get_tree tool call")
	}
	if countToolIterations(f) == 0 {
		return fmt.Errorf("the get_tree result was not fed back into the conversation")
	}
	return nil
}

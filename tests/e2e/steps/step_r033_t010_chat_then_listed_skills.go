package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T010 [BDD-RED] — Then: tellme listed the skills using its list_skills tool
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme listed the skills using its list_skills tool$`, thenListedSkillsTool)
	})
}

// thenListedSkillsTool (必查 權威狀態): the fake recorded the model's `list_skills`
// tool call AND the run fed that tool's result back into the conversation.
func thenListedSkillsTool(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	if !hasToolCall(f, "list_skills", "") {
		return fmt.Errorf("the fake recorded no list_skills tool call (requests=%d)", f.RequestCount())
	}
	if !toolResultFedBack(f, "list_skills") {
		return fmt.Errorf("the list_skills result was not fed back into the conversation")
	}
	return nil
}

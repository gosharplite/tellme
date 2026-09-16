package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T014 [BDD-RED] — Then: tellme read the skill "{name}" using its read_files tool
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme read the skill "([^"]*)" using its read_files tool$`, thenReadSkillTool)
	})
}

// thenReadSkillTool (必查 權威狀態): the fake recorded a `read_files` tool call for
// the skill {name}'s file AND the run fed that tool's result back into the
// conversation — the on-demand read reuses the EXISTING reader (NFR-002), not a
// new tool.
func thenReadSkillTool(ctx context.Context, name string) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	if !hasToolCall(f, "read_files", name) {
		return fmt.Errorf("the fake recorded no read_files tool call for the skill %q", name)
	}
	if !toolResultFedBack(f, "read_files") {
		return fmt.Errorf("the read_files result was not fed back into the conversation")
	}
	return nil
}

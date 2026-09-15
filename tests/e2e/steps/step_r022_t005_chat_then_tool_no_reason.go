package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T005 [BDD-RED] — Then: the run reported the tool call "{tool}" without a reason
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the run reported the tool call "([^"]*)" without a reason$`, thenToolCallNoReason)
	})
}

// thenToolCallNoReason (必查 呈現結果): the captured stderr carries a round-022
// `[HH:MM:SS] [Tool] <tool name>` line naming {tool} with NO reason tail (no ` - `
// separator).
func thenToolCallNoReason(ctx context.Context, tool string) error {
	sc := scenarioFrom(ctx)
	for _, tl := range toolLogs(sc.stderr) {
		if tl.name == tool && !tl.hasReason {
			return nil
		}
	}
	return fmt.Errorf("standard error carried no reasonless `[Tool]` line naming %q; stderr=%q", tool, sc.stderr)
}

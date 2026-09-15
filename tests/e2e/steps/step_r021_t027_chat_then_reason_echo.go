package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T027 (round 022 ALIGN) — Then: the run reported the reason "{reason}" for the tool call "{tool}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the run reported the reason "([^"]*)" for the tool call "([^"]*)"$`, thenReasonEcho)
	})
}

// thenReasonEcho (必查 呈現結果): the captured stderr carries a round-022 tool-loop
// log line `[HH:MM:SS] [Tool] <tool name> - <reason>` naming {tool} and carrying
// {reason} as the ` - ` tail.
func thenReasonEcho(ctx context.Context, reason, tool string) error {
	sc := scenarioFrom(ctx)
	for _, tl := range toolLogs(sc.stderr) {
		if tl.name == tool && tl.hasReason && tl.reason == reason {
			return nil
		}
	}
	return fmt.Errorf("the tool-loop log did not report reason %q for tool %q; stderr=%q", reason, tool, sc.stderr)
}

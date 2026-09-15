package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T027 [BDD-RED] — Then: the run reported the reason "{reason}" for the tool call "{tool}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the run reported the reason "([^"]*)" for the tool call "([^"]*)"$`, thenReasonEcho)
	})
}

// thenReasonEcho (必查 呈現結果): the captured stderr carries a tool-loop log line
// naming {tool} and carrying reason={reason}.
func thenReasonEcho(ctx context.Context, reason, tool string) error {
	sc := scenarioFrom(ctx)
	for _, line := range strings.Split(sc.stderr, "\n") {
		if strings.Contains(line, "[tool] "+tool) && strings.Contains(line, "reason="+reason) {
			return nil
		}
	}
	return fmt.Errorf("the tool-loop log did not report reason=%q for tool %q; stderr=%q", reason, tool, sc.stderr)
}

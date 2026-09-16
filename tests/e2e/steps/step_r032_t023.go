package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T023 [BDD-RED] — Then: the run continued past the failed MCP tool call
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the run continued past the failed MCP tool call$`, thenRunContinuedPastFailedMCP)
	})
}

// thenRunContinuedPastFailedMCP (必查 權威狀態): the failed MCP tool call was fed
// back as a recoverable tool result and the loop continued to a final answer —
// the run did NOT abort, and the frozen `the tool request failed` phrase was not
// emitted.
func thenRunContinuedPastFailedMCP(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if sc.runErr != nil {
		return fmt.Errorf("the run did not complete: %v", sc.runErr)
	}
	if sc.exitCode != 0 {
		return fmt.Errorf("the run exited %d; a recoverable MCP tool failure must not abort the run (stderr=%q)", sc.exitCode, sc.stderr)
	}
	if strings.Contains(sc.stderr, "the tool request failed") {
		return fmt.Errorf("the run emitted the terminal tool class phrase; stderr=%q", sc.stderr)
	}
	return nil
}

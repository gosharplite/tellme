package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T017 — Then: the request carried the read-tool error for "{path}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the request carried the read-tool error for "([^"]*)"$`, thenReadToolError)
	})
}

// thenReadToolError (必查 權威狀態): the fake recorded a request carrying a
// `read_files` tool result for {path} whose content is an error, fed back as a
// non-terminal result.
func thenReadToolError(ctx context.Context, path string) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	if !hasToolCall(f, "read_files", path) {
		return fmt.Errorf("no read_files tool call for %q was recorded", path)
	}
	result := lastToolResult(f)
	if result == "" {
		return fmt.Errorf("no tool result was fed back for %q", path)
	}
	lower := strings.ToLower(result)
	if !strings.Contains(lower, "error") && !strings.Contains(lower, "no such file") {
		return fmt.Errorf("the tool result for %q was not an error: %q", path, result)
	}
	return nil
}

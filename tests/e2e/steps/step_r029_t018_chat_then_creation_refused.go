package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T018 [BDD-RED] — Then: the creation is refused
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the creation is refused$`, thenCreationRefused)
	})
}

// thenCreationRefused (必查 權威狀態): the `write_file` tool result fed back is an
// error (the file already exists), so the file is left untouched.
func thenCreationRefused(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	if !hasToolCall(f, "write_file", "") {
		return fmt.Errorf("no write_file tool call was recorded")
	}
	lower := strings.ToLower(lastToolResult(f))
	if !strings.Contains(lower, "error") || !strings.Contains(lower, "exist") {
		return fmt.Errorf("the creation was not refused with an already-exists error: tool result=%q", lastToolResult(f))
	}
	return nil
}

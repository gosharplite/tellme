package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T020 [BDD-RED] — Then: the edit is refused because the block is not unique
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the edit is refused because the block is not unique$`, thenEditAmbiguous)
	})
}

// thenEditAmbiguous (必查 權威狀態): the `replace_text` tool result fed back is an
// error (the block occurs more than once).
func thenEditAmbiguous(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	if !hasToolCall(f, "replace_text", "") {
		return fmt.Errorf("no replace_text tool call was recorded")
	}
	lower := strings.ToLower(lastToolResult(f))
	if !strings.Contains(lower, "error") || !strings.Contains(lower, "unique") {
		return fmt.Errorf("the edit was not refused as a non-unique block: tool result=%q", lastToolResult(f))
	}
	return nil
}

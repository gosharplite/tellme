package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T013 — Then: the active session history holds no exchanges
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the active session history holds no exchanges$`, thenActiveHistoryEmpty)
	})
}

// thenActiveHistoryEmpty (必查 權威狀態): the active history file holds no exchange
// lines.
func thenActiveHistoryEmpty(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	entries, err := readHistoryEntries(sc.historyFilePath())
	if err != nil {
		return fmt.Errorf("read session history: %w", err)
	}
	if len(entries) != 0 {
		return fmt.Errorf("the active history holds %d exchanges, want none; %+v", len(entries), entries)
	}
	return nil
}

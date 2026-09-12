package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T008 — Then: tellme stored the exchange "{prompt}" and "{answer}" in the session history
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme stored the exchange "([^"]*)" and "([^"]*)" in the session history$`, thenStoredExchange)
	})
}

// thenStoredExchange (必查 權威狀態): the active history file holds a line whose
// prompt and answer are {prompt} and {answer}.
func thenStoredExchange(ctx context.Context, prompt, answer string) error {
	sc := scenarioFrom(ctx)
	entries, err := readHistoryEntries(sc.historyFilePath())
	if err != nil {
		return fmt.Errorf("read session history: %w", err)
	}
	for _, e := range entries {
		if e.Prompt == prompt && e.Answer == answer {
			return nil
		}
	}
	return fmt.Errorf("the session history did not store %q/%q; active=%+v", prompt, answer, entries)
}

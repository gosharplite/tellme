package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T014 — Then: the archived session history holds the exchange "{prompt}" and "{answer}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the archived session history holds the exchange "([^"]*)" and "([^"]*)"$`, thenArchivedHoldsExchange)
	})
}

// thenArchivedHoldsExchange (必查 權威狀態): the archive file holds a line whose
// prompt and answer are {prompt} and {answer}.
func thenArchivedHoldsExchange(ctx context.Context, prompt, answer string) error {
	sc := scenarioFrom(ctx)
	entries, err := readHistoryEntries(sc.historyArchivePath())
	if err != nil {
		return fmt.Errorf("read session archive: %w", err)
	}
	for _, e := range entries {
		if e.Prompt == prompt && e.Answer == answer {
			return nil
		}
	}
	return fmt.Errorf("the archive did not hold %q/%q; archived=%+v", prompt, answer, entries)
}

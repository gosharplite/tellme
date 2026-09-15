package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T024 — Then: the shared prompt log records the prompt "{prompt}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the shared prompt log records the prompt "([^"]*)"$`, thenSharedLogRecords)
	})
}

// thenSharedLogRecords (必查 權威狀態): the shared log holds {prompt}. 工作區: reads
// ~/.tellme/global_prompts.jsonl.
func thenSharedLogRecords(ctx context.Context, prompt string) error {
	sc := scenarioFrom(ctx)
	prompts, err := readPromptLog(sc)
	if err != nil {
		return err
	}
	for _, p := range prompts {
		if p == prompt {
			return nil
		}
	}
	return fmt.Errorf("the shared prompt log does not record %q; got %v", prompt, prompts)
}

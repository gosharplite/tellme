package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// R028 — Then: the shared prompt log does not hold "{prompt}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the shared prompt log does not hold "([^"]*)"$`, thenSharedLogNotHolds)
	})
}

// thenSharedLogNotHolds (必查 權威狀態): the user-global shared log holds no record
// whose prompt is {prompt} — a present shared log is not overwritten by a
// carry-over. 工作區: reads ~/.tellme/global_prompts.jsonl.
func thenSharedLogNotHolds(ctx context.Context, prompt string) error {
	sc := scenarioFrom(ctx)
	prompts, err := readPromptLog(sc)
	if err != nil {
		return err
	}
	for _, p := range prompts {
		if p == prompt {
			return fmt.Errorf("the shared prompt log should not hold %q; got %v", prompt, prompts)
		}
	}
	return nil
}

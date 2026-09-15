package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// R028 — Then: the shared prompt log also holds the earlier prompt "{prompt}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the shared prompt log also holds the earlier prompt "([^"]*)"$`, thenSharedLogAlsoHolds)
	})
}

// thenSharedLogAlsoHolds (必查 權威狀態): the user-global shared log holds the
// carried-over {prompt} (the round-028 first-use seed). 工作區: reads
// ~/.tellme/global_prompts.jsonl.
func thenSharedLogAlsoHolds(ctx context.Context, prompt string) error {
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
	return fmt.Errorf("the shared prompt log does not hold the carried-over prompt %q; got %v", prompt, prompts)
}

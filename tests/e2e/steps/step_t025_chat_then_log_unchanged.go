package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T025 — Then: the shared prompt log still holds only "{prompt}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the shared prompt log still holds only "([^"]*)"$`, thenSharedLogStillOnly)
	})
}

// thenSharedLogStillOnly (必查 權威狀態): the shared log holds exactly the arranged
// prompt {prompt} and nothing more — a non-`-i` run appended nothing. 工作區:
// reads $TELL_ME_HOME/output/global_prompts.jsonl.
func thenSharedLogStillOnly(ctx context.Context, prompt string) error {
	sc := scenarioFrom(ctx)
	prompts, err := readPromptLog(sc)
	if err != nil {
		return err
	}
	if len(prompts) != 1 || prompts[0] != prompt {
		return fmt.Errorf("the shared prompt log should hold only %q; got %v", prompt, prompts)
	}
	return nil
}

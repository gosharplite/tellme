package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// R064 — Then: the interactive prompt does not offer the recent prompt "{prompt}".
//
// Round 064 (TD-1): the negative counterpart of the positive presence assertion
// (`the interactive prompt offers the recent prompt "{prompt}"` — the sentence
// owned by `step_t019_chat_then_offers_recent_prompt.go`) — the missing half of
// the surface-cap acceptance Rule. With more matching prompts than the cap, an
// item dropped by the cap must be absent from the rendered suggestion list
// (absence, not count — robust against TUI repaint frames).
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the interactive prompt does not offer the recent prompt "([^"]*)"$`, thenDoesNotOfferRecentPrompt)
	})
}

func thenDoesNotOfferRecentPrompt(ctx context.Context, prompt string) error {
	sc := scenarioFrom(ctx)
	out := renderedOutput(sc)
	if strings.Contains(out, prompt) {
		return fmt.Errorf("the interactive prompt offered the recent prompt %q, want it dropped by the surface cap; output=%q", prompt, out)
	}
	return nil
}

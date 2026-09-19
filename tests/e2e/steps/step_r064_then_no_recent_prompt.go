package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// R064 — Then: the interactive prompt does not offer the recent prompt "{prompt}".
//
// Round 064 (TD-1): the negative counterpart of T019's presence assertion — the
// missing half of the surface-cap acceptance Rule. With more matching prompts
// than the cap, an item dropped by the cap must be absent from the rendered
// suggestion list (absence, not count — robust against TUI repaint frames).
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the interactive prompt does not offer the recent prompt "([^"]*)"$`, thenDoesNotOfferRecentPrompt)
	})
}

func thenDoesNotOfferRecentPrompt(ctx context.Context, prompt string) error {
	sc := scenarioFrom(ctx)
	if strings.Contains(renderedOutput(sc), prompt) {
		return fmt.Errorf("the interactive prompt offered the recent prompt %q, want it dropped by the surface cap; output=%q", prompt, renderedOutput(sc))
	}
	return nil
}

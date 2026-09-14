package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T019 — Then: the interactive prompt offers the recent prompt "{prompt}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the interactive prompt offers the recent prompt "([^"]*)"$`, thenOffersRecentPrompt)
	})
}

// thenOffersRecentPrompt (必查 呈現結果): the captured output contains {prompt} as
// a suggestion (presence, not exact ANSI bytes — round-015 chat/dsl.md).
func thenOffersRecentPrompt(ctx context.Context, prompt string) error {
	sc := scenarioFrom(ctx)
	if !strings.Contains(renderedOutput(sc), prompt) {
		return fmt.Errorf("the interactive prompt did not offer the recent prompt %q; output=%q", prompt, renderedOutput(sc))
	}
	return nil
}

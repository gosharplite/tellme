package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T022 — Then: the interactive prompt reports the active provider "{provider}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the interactive prompt reports the active provider "([^"]*)"$`, thenReportsActiveProvider)
	})
}

// thenReportsActiveProvider (必查 呈現結果): the captured output contains the active
// provider/model {provider} (the dashboard header) — presence.
func thenReportsActiveProvider(ctx context.Context, provider string) error {
	sc := scenarioFrom(ctx)
	if !strings.Contains(renderedOutput(sc), provider) {
		return fmt.Errorf("the interactive prompt did not report the active provider %q; output=%q", provider, renderedOutput(sc))
	}
	return nil
}

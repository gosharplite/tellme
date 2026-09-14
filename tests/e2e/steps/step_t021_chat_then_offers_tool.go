package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T021 — Then: the interactive prompt offers the available tool "{tool}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the interactive prompt offers the available tool "([^"]*)"$`, thenOffersTool)
	})
}

// thenOffersTool (必查 呈現結果): the captured output contains the registered tool
// name {tool} as a suggestion (presence). 來源: tellme's registered tool registry.
func thenOffersTool(ctx context.Context, tool string) error {
	sc := scenarioFrom(ctx)
	if !strings.Contains(renderedOutput(sc), tool) {
		return fmt.Errorf("the interactive prompt did not offer the available tool %q; output=%q", tool, renderedOutput(sc))
	}
	return nil
}

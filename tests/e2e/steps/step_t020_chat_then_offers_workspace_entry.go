package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T020 — Then: the interactive prompt offers the workspace entry "{entry}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the interactive prompt offers the workspace entry "([^"]*)"$`, thenOffersWorkspaceEntry)
	})
}

// thenOffersWorkspaceEntry (必查 呈現結果): the captured output contains {entry} as
// a suggestion (presence).
func thenOffersWorkspaceEntry(ctx context.Context, entry string) error {
	sc := scenarioFrom(ctx)
	if !strings.Contains(renderedOutput(sc), entry) {
		return fmt.Errorf("the interactive prompt did not offer the workspace entry %q; output=%q", entry, renderedOutput(sc))
	}
	return nil
}

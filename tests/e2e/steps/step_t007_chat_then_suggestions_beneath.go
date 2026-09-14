package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T007 — Then: the interactive prompt lists suggestions beneath the editor
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the interactive prompt lists suggestions beneath the editor$`, thenSuggestionsBeneath)
	})
}

// thenSuggestionsBeneath (必查 呈現結果): the captured output carries the
// suggestion-list header beneath the editor frame.
func thenSuggestionsBeneath(ctx context.Context) error {
	if out := renderedOutput(scenarioFrom(ctx)); !tuiSuggestionsBeneath(out) {
		return fmt.Errorf("the interactive prompt did not list suggestions beneath the editor; output=%q", out)
	}
	return nil
}

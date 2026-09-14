package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T013 — Then: the interactive prompt holds the accepted suggestion "{text}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the interactive prompt holds the accepted suggestion "([^"]*)"$`, thenHoldsAcceptedSuggestion)
	})
}

// thenHoldsAcceptedSuggestion (必查 呈現結果): the captured output carries {text} in
// an EDITOR row (the accepted suggestion was inserted — not merely shown in the
// suggestion list, which would make the assertion vacuous, F2).
func thenHoldsAcceptedSuggestion(ctx context.Context, text string) error {
	out := renderedOutput(scenarioFrom(ctx))
	if !tuiEditorRowContains(out, unescapeText(text)) {
		return fmt.Errorf("the editor did not hold the accepted suggestion %q; output=%q", text, out)
	}
	return nil
}

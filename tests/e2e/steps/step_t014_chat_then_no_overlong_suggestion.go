package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T014 — Then: no suggestion offered to the operator spans more than three lines
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^no suggestion offered to the operator spans more than three lines$`, thenNoOverlongSuggestion)
	})
}

// thenNoOverlongSuggestion (必查 呈現結果): the rendered suggestion list carries
// no wrapped continuation row (an entry over three lines is dropped, FR-006).
func thenNoOverlongSuggestion(ctx context.Context) error {
	if out := renderedOutput(scenarioFrom(ctx)); tuiOverlongSuggestionRow(out) {
		return fmt.Errorf("a suggestion spanned more than three lines; output=%q", out)
	}
	return nil
}

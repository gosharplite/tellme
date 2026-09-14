package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T011 — Then: the interactive prompt keeps the typed text "{text}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the interactive prompt keeps the typed text "([^"]*)"$`, thenKeepsTypedText)
	})
}

// thenKeepsTypedText (必查 呈現結果): the captured output carries the typed text in
// an EDITOR row (not merely in the suggestion list). A multi-line value (decoded
// via unescapeText) is asserted line-by-line.
func thenKeepsTypedText(ctx context.Context, text string) error {
	out := renderedOutput(scenarioFrom(ctx))
	for _, part := range strings.Split(unescapeText(text), "\n") {
		if part != "" && !tuiEditorRowContains(out, part) {
			return fmt.Errorf("the interactive prompt did not keep the typed text %q (missing %q in the editor); output=%q", text, part, out)
		}
	}
	return nil
}

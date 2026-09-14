package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T015 — Then: the interactive prompt shows no line numbers
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the interactive prompt shows no line numbers$`, thenNoLineNumbers)
	})
}

// thenNoLineNumbers (必查 呈現結果): the editor renders no line-number gutter.
func thenNoLineNumbers(ctx context.Context) error {
	if out := renderedOutput(scenarioFrom(ctx)); tuiHasLineNumberGutter(out) {
		return fmt.Errorf("the editor rendered line numbers; output=%q", out)
	}
	return nil
}

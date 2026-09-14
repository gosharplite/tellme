package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T006 — Then: the interactive prompt is framed around a multi-line editor
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the interactive prompt is framed around a multi-line editor$`, thenPromptFramed)
	})
}

// thenPromptFramed (必查 呈現結果): the captured output draws the editor border.
func thenPromptFramed(ctx context.Context) error {
	if out := renderedOutput(scenarioFrom(ctx)); !tuiHasBorder(out) {
		return fmt.Errorf("the interactive prompt was not framed around a multi-line editor; output=%q", out)
	}
	return nil
}

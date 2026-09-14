package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/internal/cli"
)

// T018 — Then: the interactive prompt is not shown
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the interactive prompt is not shown$`, thenInteractivePromptNotShown)
	})
}

// thenInteractivePromptNotShown (必查 呈現結果): neither the captured standard
// output nor standard error carries the interactive-prompt announcement — the
// prompt must not render on a non-terminal input. The literal is single-sourced
// from the product (cli.TUIHint).
func thenInteractivePromptNotShown(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if strings.Contains(sc.stdout, cli.TUIHint) || strings.Contains(sc.stderr, cli.TUIHint) {
		return fmt.Errorf("the interactive prompt was shown; stdout=%q stderr=%q", sc.stdout, sc.stderr)
	}
	return nil
}

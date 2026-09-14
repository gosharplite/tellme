package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/internal/cli"
)

// T017 — Then: the interactive prompt is shown
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the interactive prompt is shown$`, thenInteractivePromptShown)
	})
}

// thenInteractivePromptShown (必查 呈現結果): the captured standard error carries
// the interactive-prompt announcement. The literal is single-sourced from the
// product (cli.TUIHint), so a product-side text change cannot make the guard
// vacuous (mirroring the round-012 TD1 pattern); it must NOT be on stdout.
func thenInteractivePromptShown(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if !strings.Contains(sc.stderr, cli.TUIHint) {
		return fmt.Errorf("the interactive prompt announcement was not reported on stderr; stderr=%q", sc.stderr)
	}
	if strings.Contains(sc.stdout, cli.TUIHint) {
		return fmt.Errorf("the interactive prompt announcement leaked onto standard output; stdout=%q", sc.stdout)
	}
	return nil
}

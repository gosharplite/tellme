package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/harness"
)

// T018 — Then: tellme prints the provider's answer "{answer}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme prints the provider's answer "([^"]*)"$`, thenPrintsAnswer)
	})
}

// thenPrintsAnswer (必查 呈現結果): the captured standard output contains the
// provider's answer and it is not swallowed or emitted only on stderr. The
// default output is Markdown-rendered (round 006), so the assertion compares the
// VISIBLE text — the renderer's ANSI escapes (terminal/profile-dependent) are
// stripped first; under -r the output is already plain, so this is a no-op there.
func thenPrintsAnswer(ctx context.Context, answer string) error {
	sc := scenarioFrom(ctx)
	if !strings.Contains(harness.StripANSI(sc.stdout), answer) {
		return fmt.Errorf("stdout %q does not contain the provider's answer %q", sc.stdout, answer)
	}
	return nil
}

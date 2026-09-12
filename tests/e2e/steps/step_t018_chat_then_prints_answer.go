package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T018 — Then: tellme prints the provider's answer "{answer}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme prints the provider's answer "([^"]*)"$`, thenPrintsAnswer)
	})
}

// thenPrintsAnswer (必查 呈現結果): stdout contains the provider's answer and it
// is not swallowed or emitted only on stderr.
func thenPrintsAnswer(ctx context.Context, answer string) error {
	sc := scenarioFrom(ctx)
	if !strings.Contains(sc.stdout, answer) {
		return fmt.Errorf("stdout %q does not contain the provider's answer %q", sc.stdout, answer)
	}
	return nil
}

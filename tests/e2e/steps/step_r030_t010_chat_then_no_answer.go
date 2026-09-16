package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T010 [BDD-RED] — Then: tellme prints no answer
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme prints no answer$`, thenPrintsNoAnswer)
	})
}

// thenPrintsNoAnswer (必查 呈現結果): the captured standard output is empty — the
// failed turn emitted no answer bytes (a cut-off answer must never be printed as
// the answer).
func thenPrintsNoAnswer(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if strings.TrimSpace(sc.stdout) != "" {
		return fmt.Errorf("expected no answer on stdout, but got %q", sc.stdout)
	}
	return nil
}

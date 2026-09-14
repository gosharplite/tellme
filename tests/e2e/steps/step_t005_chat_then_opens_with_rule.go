package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T005 — Then: the turn opens with a horizontal rule
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the turn opens with a horizontal rule$`, thenTurnOpensWithRule)
	})
}

// thenTurnOpensWithRule (必查 呈現結果): the diagnostic stream carries a line that
// is exactly the 80-column `─` rule, and it is NOT written to standard output.
func thenTurnOpensWithRule(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if !hasTurnRule(sc.stderr) {
		return fmt.Errorf("the diagnostic stream carried no 80-column horizontal rule; stderr=%q", sc.stderr)
	}
	if hasTurnRule(sc.stdout) {
		return fmt.Errorf("the horizontal rule leaked to standard output; stdout=%q", sc.stdout)
	}
	return nil
}

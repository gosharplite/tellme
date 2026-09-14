package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T010 — Then: the run reports the cost of the request, the turn, and the session
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the run reports the cost of the request, the turn, and the session$`, thenCostReported)
	})
}

// thenCostReported (必查 呈現結果): the captured standard error carries a summary
// line matching `╰─⠿ Ready ($… $… $… - M: … H: … O: … - …%)`; it must NOT be on
// standard output.
func thenCostReported(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if !hasReadyLine(sc.stderr) {
		return fmt.Errorf("standard error carried no Ready summary line; stderr=%q", sc.stderr)
	}
	if hasReadyLine(sc.stdout) {
		return fmt.Errorf("the Ready summary must not be written to standard output; stdout=%q", sc.stdout)
	}
	return nil
}

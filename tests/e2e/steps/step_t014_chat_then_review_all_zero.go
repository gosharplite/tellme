package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T014 [BDD-RED] — Then: the review shows every tool with no uses
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the review shows every tool with no uses$`, thenReviewEveryToolZero)
	})
}

// thenReviewEveryToolZero (必查 呈現結果): the report lists EVERY recordable tool
// (the union of the base and capability-gated sets — round 062 / PR #129 fold
// F-062-1), each with zero invocations — so a tool add/remove cannot pass
// vacuously (round-026 review F8).
func thenReviewEveryToolZero(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	names := recordableToolNames()
	if len(names) == 0 {
		return fmt.Errorf("the recordable set enumerates no tools")
	}
	for _, n := range names {
		counts, found := parseReportLine(sc.stdout, n)
		if !found {
			return fmt.Errorf("the report omits the registered tool %q; stdout=%q", n, sc.stdout)
		}
		if counts.total != 0 {
			return fmt.Errorf("the report shows %q with %d invocations, want 0", n, counts.total)
		}
	}
	return nil
}

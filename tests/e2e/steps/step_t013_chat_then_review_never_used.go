package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T013 [BDD-RED] — Then: the review shows the tool "{tool}" was never used
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the review shows the tool "([^"]*)" was never used$`, thenReviewNeverUsed)
	})
}

// thenReviewNeverUsed (必查 呈現結果): the report lists {tool} with zero
// invocations (a registered-but-unused tool must not be omitted).
func thenReviewNeverUsed(ctx context.Context, tool string) error {
	sc := scenarioFrom(ctx)
	counts, found := parseReportLine(sc.stdout, tool)
	if !found {
		return fmt.Errorf("the report omits the registered tool %q; stdout=%q", tool, sc.stdout)
	}
	if counts.total != 0 {
		return fmt.Errorf("the report shows %q with %d invocations, want 0 (never used)", tool, counts.total)
	}
	return nil
}

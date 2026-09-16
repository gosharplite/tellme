package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T013 [BDD-RED] — Then: the listing reports that no skills are available
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the listing reports that no skills are available$`, thenListingReportsNoSkills)
	})
}

// thenListingReportsNoSkills (必查 權威狀態): the `list_skills` result reports an empty
// catalog — not an error, and not a non-empty list. The tool renders a clear
// "no skills" line for an empty catalog.
func thenListingReportsNoSkills(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	result := lastToolResult(f)
	if !strings.Contains(strings.ToLower(result), "no skills") {
		return fmt.Errorf("the listing does not report an empty catalog; result=%q", result)
	}
	return nil
}

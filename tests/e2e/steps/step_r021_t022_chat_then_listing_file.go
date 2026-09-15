package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T022 [BDD-RED] — Then: the listing shows "{name}" as a file
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the listing shows "([^"]*)" as a file$`, thenListingFile)
	})
}

// thenListingFile (必查 權威狀態): the list_files result carries a "[f] {name}" line.
func thenListingFile(ctx context.Context, name string) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	if !listingHas(lastToolResult(f), "f", name) {
		return fmt.Errorf("the listing did not show %q as a file; result=%q", name, lastToolResult(f))
	}
	return nil
}

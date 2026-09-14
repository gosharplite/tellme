package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T023 [BDD-RED] — Then: the listing shows "{name}" as a folder
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the listing shows "([^"]*)" as a folder$`, thenListingFolder)
	})
}

// thenListingFolder (必查 權威狀態): the list_files result carries a "[d] {name}" line.
func thenListingFolder(ctx context.Context, name string) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	if !listingHas(lastToolResult(f), "d", name) {
		return fmt.Errorf("the listing did not show %q as a folder; result=%q", name, lastToolResult(f))
	}
	return nil
}

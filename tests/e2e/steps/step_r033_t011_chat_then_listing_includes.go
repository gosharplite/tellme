package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T011 [BDD-RED] — Then: the listing includes the skill "{name}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the listing includes the skill "([^"]*)"$`, thenListingIncludesSkill)
	})
}

// thenListingIncludesSkill (必查 權威狀態): the `list_skills` result carries an entry
// for a skill named {name}.
func thenListingIncludesSkill(ctx context.Context, name string) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	result := lastToolResult(f)
	if !listingHasSkill(result, name) {
		return fmt.Errorf("the listing does not include the skill %q; result=%q", name, result)
	}
	return nil
}

package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T028 [BDD-RED] — Then: the tree does not show "{entry}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the tree does not show "([^"]*)"$`, thenTreeAbsent)
	})
}

// thenTreeAbsent (必查 權威狀態): the get_tree result carries no connector line for
// {entry} — an entry beyond the depth bound must not appear.
func thenTreeAbsent(ctx context.Context, entry string) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	if treeLists(lastToolResult(f), entry) {
		return fmt.Errorf("the tree unexpectedly shows %q; result=%q", entry, lastToolResult(f))
	}
	return nil
}

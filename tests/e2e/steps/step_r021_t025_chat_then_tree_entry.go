package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T025 [BDD-RED] — Then: the tree shows "{entry}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the tree shows "([^"]*)"$`, thenTreeEntry)
	})
}

// thenTreeEntry (必查 權威狀態): the get_tree result carries a connector line whose
// entry name is {entry}.
func thenTreeEntry(ctx context.Context, entry string) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	if !treeLists(lastToolResult(f), entry) {
		return fmt.Errorf("the tree does not show %q; result=%q", entry, lastToolResult(f))
	}
	return nil
}

package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T029 [BDD-RED] — Then: the tree does not descend into "{name}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the tree does not descend into "([^"]*)"$`, thenTreeNoDescend)
	})
}

// thenTreeNoDescend (必查 權威狀態): the get_tree result lists {name} but carries no
// connector line nested under it (e.g. `.git` is listed, never recursed).
func thenTreeNoDescend(ctx context.Context, name string) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	result := lastToolResult(f)
	if !treeLists(result, name) {
		return fmt.Errorf("the tree does not list %q; result=%q", name, result)
	}
	if !treeNotDescends(result, name) {
		return fmt.Errorf("the tree descends into %q; result=%q", name, result)
	}
	return nil
}

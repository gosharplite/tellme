package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T017 [BDD-RED] — Then: the run made a single read_files request carrying "{path_a}" and "{path_b}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the run made a single read_files request carrying "([^"]*)" and "([^"]*)"$`, thenSingleReadTwo)
	})
}

// thenSingleReadTwo (必查 權威狀態): the fake recorded exactly one read_files tool
// call and its `filepaths` are {path_a} and {path_b} (in order).
func thenSingleReadTwo(ctx context.Context, pathA, pathB string) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	calls := readFilesCalls(f)
	if len(calls) != 1 {
		return fmt.Errorf("expected exactly one read_files call, got %d", len(calls))
	}
	if len(calls[0]) != 2 || calls[0][0] != pathA || calls[0][1] != pathB {
		return fmt.Errorf("read_files filepaths = %v; want [%q %q]", calls[0], pathA, pathB)
	}
	return nil
}

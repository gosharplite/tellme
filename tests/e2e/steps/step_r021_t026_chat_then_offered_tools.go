package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T026 [BDD-RED] — Then: the request offered exactly the reader tools
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the request offered exactly the reader tools$`, thenOfferedTools)
	})
}

// thenOfferedTools (必查 權威狀態): the recorded request offered exactly the three
// filesystem reader tools and no other (the summarisation tool must not appear).
func thenOfferedTools(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	names := f.ToolNamesAt(-1)
	want := map[string]bool{"list_files": true, "read_files": true, "get_tree": true}
	if len(names) != len(want) {
		return fmt.Errorf("offered tools = %v; want exactly list_files, read_files, get_tree", names)
	}
	for _, n := range names {
		if !want[n] {
			return fmt.Errorf("offered unexpected tool %q (want only list_files, read_files, get_tree)", n)
		}
	}
	return nil
}

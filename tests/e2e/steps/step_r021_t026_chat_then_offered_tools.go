package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T012 [BDD-ALIGN] — Then: the request offered exactly the agent tools
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the request offered exactly the agent tools$`, thenOfferedTools)
	})
}

// thenOfferedTools (必查 權威狀態): the recorded request offered exactly the four
// agent tools (the three readers plus execute_command) and no other (the
// summarisation tool must not appear; no pipe_commands).
func thenOfferedTools(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	names := f.ToolNamesAt(-1)
	want := map[string]bool{"list_files": true, "read_files": true, "get_tree": true, "execute_command": true}
	if len(names) != len(want) {
		return fmt.Errorf("offered tools = %v; want exactly list_files, read_files, get_tree, execute_command", names)
	}
	for _, n := range names {
		if !want[n] {
			return fmt.Errorf("offered unexpected tool %q", n)
		}
	}
	return nil
}

package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T005 [BDD-ALIGN] — Then: the request offered exactly the agent tools
//
// Round 029 MODIFY: the offered set grew from four tools (the three readers +
// execute_command) to six (three readers + the write pair write_file /
// replace_text + execute_command).
//
// Round 033 MODIFY (T004): the offered set grew to seven — the reader trio, the
// write pair, execute_command, and the read-only list_skills tool. The expected
// set is single-sourced from the shared registeredToolNames() enumerator, so no
// count is hand-copied. Aligned to the latest `chat/dsl.md` row.
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the request offered exactly the agent tools$`, thenOfferedTools)
	})
}

// thenOfferedTools (必查 權威狀態): the recorded request offered exactly the tools
// the LIVE registry enumerates, and no other (the summarisation tool and
// pipe_commands must not appear).
//
// Round-029 PR #61 review finding 7 + implementation review finding 3: the
// expected set is NOT hand-copied. It is single-sourced from the shared
// registeredToolNames() enumerator (the same constructors cli.newToolRegistry
// uses, also used by the round-026 all-zero Then) — one owner for the set, in
// process, with no subprocess/log dependency.
func thenOfferedTools(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	offered := f.ToolNamesAt(-1)
	want := registeredToolNames()
	if len(want) == 0 {
		return fmt.Errorf("the live registry enumerates no tools")
	}
	if len(offered) != len(want) {
		return fmt.Errorf("offered tools = %v; want exactly the live registry %v", offered, want)
	}
	set := make(map[string]bool, len(want))
	for _, n := range want {
		set[n] = true
	}
	for _, n := range offered {
		if !set[n] {
			return fmt.Errorf("offered unexpected tool %q (not in the live registry %v)", n, want)
		}
	}
	return nil
}

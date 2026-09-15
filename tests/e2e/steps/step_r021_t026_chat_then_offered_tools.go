package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/harness"
)

// T005 [BDD-ALIGN] — Then: the request offered exactly the agent tools
//
// Round 029 MODIFY: the offered set grew from four tools (the three readers +
// execute_command) to six (three readers + the write pair write_file /
// replace_text + execute_command). Aligned to the latest `chat/dsl.md` row.
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the request offered exactly the agent tools$`, thenOfferedTools)
	})
}

// thenOfferedTools (必查 權威狀態): the recorded request offered exactly the tools
// the LIVE production registry offers, and no other (the summarisation tool and
// pipe_commands must not appear).
//
// Round-029 PR #61 review finding 7: the expected set is NOT hand-copied a third
// time. It is derived from `tellme --tool-usage` — the offline, --version-class
// report that lists every tool the live registry offers (round-026) — so adding
// or removing a tool cannot drift from the `chat/dsl.md` 集合 or the acceptance
// prose.
func thenOfferedTools(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	offered := f.ToolNamesAt(-1)
	want := liveRegistryToolNames(sc)
	if len(want) == 0 {
		return fmt.Errorf("the live registry enumerated no tools (the --tool-usage report listed no tool lines)")
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

// liveRegistryToolNames runs `tellme --tool-usage` (the --version-class offline
// report listing every tool the live registry offers) and parses the tool names,
// so the expected offered set tracks the production registry rather than a
// hand-copied constant. Each report line has the shape `<tool>: total=… ok=…`.
func liveRegistryToolNames(sc *scenarioContext) []string {
	res := harness.RunIn(sc.workDir, []string{"--tool-usage"}, sc.runEnv(), sc.unsetNames())
	var names []string
	for _, line := range strings.Split(res.Stdout, "\n") {
		line = strings.TrimRight(line, "\r")
		if i := strings.Index(line, ": total="); i > 0 {
			names = append(names, line[:i])
		}
	}
	return names
}

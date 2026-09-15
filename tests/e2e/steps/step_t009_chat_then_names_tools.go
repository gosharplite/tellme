package steps

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/cucumber/godog"
)

// T003 [BDD-ALIGN] (round 025) — Then: the progress spinner names the first tool
// and counts the remaining tools.
//
// Round 025 renamed this from round-019's `the progress spinner names every tool
// it is running`: the several-tool label is now BOUNDED
// (`Executing tools [<first> and <N-1> more]...`), so it no longer enumerates
// every name (issue #55).
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the progress spinner names the first tool and counts the remaining tools$`, thenSpinnerNamesBoundedTools)
	})
}

// reBoundedTools captures the first name and the remaining count of the bounded
// several-tool status form.
var reBoundedTools = regexp.MustCompile(`Executing tools \[([^\]]+?) and ([0-9]+) more\]\.\.\.`)

// thenSpinnerNamesBoundedTools (必查 / 呈現結果): the captured stderr carries a
// spinner line whose status is the BOUNDED several-tool form
// `Executing tools [<first> and <N-1> more]...` (the first tool name + the count
// of the rest). 不該發生: the several-tool status must not enumerate every name
// (the retired round-019 form), must not collapse to the single-tool form, and
// must not omit the count.
func thenSpinnerNamesBoundedTools(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	m := reBoundedTools.FindStringSubmatch(sc.stderr)
	if m == nil {
		return fmt.Errorf("the spinner did not use the bounded several-tool form: stderr=%q", sc.stderr)
	}
	if m[1] == "" {
		return fmt.Errorf("the bounded several-tool form named an empty first tool: %q", m[0])
	}
	if sc.scriptedTool != "" && m[1] != sc.scriptedTool {
		return fmt.Errorf("the bounded form named the first tool %q, want %q: %q", m[1], sc.scriptedTool, m[0])
	}
	n, err := strconv.Atoi(m[2])
	if err != nil || n < 1 {
		return fmt.Errorf("the bounded several-tool form carried a bad remaining-count %q", m[2])
	}
	// The scenario scripts exactly two tools, so the count of the rest is 1.
	if n != 1 {
		return fmt.Errorf("the bounded form counted %d remaining tool(s), want 1 (two tools): %q", n, m[0])
	}
	return nil
}

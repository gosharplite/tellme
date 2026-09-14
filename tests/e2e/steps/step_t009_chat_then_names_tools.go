package steps

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/cucumber/godog"
)

// T009 [BDD-RED] — Then: the progress spinner names every tool it is running.
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the progress spinner names every tool it is running$`, thenSpinnerNamesTools)
	})
}

// reT009ExecutingTools captures the inner list of the several-tool status form.
var reT009ExecutingTools = regexp.MustCompile(`Executing tools \[([^\]]+)\]\.\.\.`)

// thenSpinnerNamesTools (必查 / 呈現結果): the captured stderr carries a spinner line
// whose status is the several-tool form `Executing tools [<a>, <b>]...` naming
// every tool the run is executing. 不該發生: the several-tool status must not
// collapse to the single-tool form nor omit a tool.
func thenSpinnerNamesTools(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	m := reT009ExecutingTools.FindStringSubmatch(sc.stderr)
	if m == nil {
		return fmt.Errorf("the spinner did not use the several-tool form: stderr=%q", sc.stderr)
	}
	names := strings.Split(m[1], ", ")
	if len(names) < 2 {
		return fmt.Errorf("the several-tool form named %d tool(s), want >= 2: %q", len(names), m[1])
	}
	for _, n := range names {
		if n == "" {
			return fmt.Errorf("the several-tool form named an empty tool: %q", m[1])
		}
		if sc.scriptedTool != "" && n != sc.scriptedTool {
			return fmt.Errorf("the several-tool form named %q, want every tool %q: %q", n, sc.scriptedTool, m[1])
		}
	}
	return nil
}

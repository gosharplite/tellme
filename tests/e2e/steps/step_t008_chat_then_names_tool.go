package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T008 [BDD-RED] — Then: the progress spinner names the tool it is running.
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the progress spinner names the tool it is running$`, thenSpinnerNamesTool)
	})
}

// thenSpinnerNamesTool (必查 / 呈現結果): the captured stderr carries a spinner line
// whose status is the single-tool form `Executing [<tool>]...` naming the tool the
// run is executing (the scenario's scripted tool). 不該發生: the tool-phase spinner
// must not omit the tool name.
func thenSpinnerNamesTool(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	tool := sc.scriptedTool
	if tool == "" {
		return fmt.Errorf("the scenario scripted no tool to name")
	}
	want := "Executing [" + tool + "]..."
	if !strings.Contains(sc.stderr, want) {
		return fmt.Errorf("the spinner did not name the tool %q (want %q): stderr=%q", tool, want, sc.stderr)
	}
	return nil
}

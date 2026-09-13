package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T016 — Then: the run reported the tool call "{tool}" on its diagnostic output
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the run reported the tool call "([^"]*)" on its diagnostic output$`, thenReportedToolCall)
	})
}

// thenReportedToolCall (必查 呈現結果): the captured standard error carries a
// tool-loop log line naming {tool}, and the log is NOT written to standard
// output. The assertion keys on the DSL contract only — "a tool-loop log line
// naming {tool}" (specs/truth/features/cli/chat/dsl.md) — deliberately WITHOUT
// coupling to any specific log token (e.g. `[tool]`), which truth does not pin
// (T025 review issue #1).
func thenReportedToolCall(ctx context.Context, tool string) error {
	sc := scenarioFrom(ctx)
	if !strings.Contains(sc.stderr, tool) {
		return fmt.Errorf("standard error carried no tool-loop log line naming %q; stderr=%q", tool, sc.stderr)
	}
	if strings.Contains(sc.stdout, tool) {
		return fmt.Errorf("the tool-loop activity leaked onto standard output: %q", sc.stdout)
	}
	return nil
}

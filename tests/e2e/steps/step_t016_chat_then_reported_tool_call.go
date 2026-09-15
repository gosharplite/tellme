package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T016 (round 022 ALIGN) — Then: the run reported the tool call "{tool}" on its diagnostic output
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the run reported the tool call "([^"]*)" on its diagnostic output$`, thenReportedToolCall)
	})
}

// thenReportedToolCall (必查 呈現結果): the captured standard error carries a
// tool-loop log line of the round-022 `[HH:MM:SS] [Tool] <tool name>` shape naming
// {tool}; the line must NOT echo the call's raw arguments/result, and the tool
// activity must NOT be written to standard output.
func thenReportedToolCall(ctx context.Context, tool string) error {
	sc := scenarioFrom(ctx)
	for _, line := range strings.Split(sc.stderr, "\n") {
		tl, ok := parseToolLog(line)
		if !ok || tl.name != tool {
			continue
		}
		if strings.Contains(line, "arguments=") || strings.Contains(line, "result=") {
			return fmt.Errorf("the tool-loop log line echoed the raw arguments/result: %q", line)
		}
		if strings.Contains(sc.stdout, tool) {
			return fmt.Errorf("the tool-loop activity leaked onto standard output: %q", sc.stdout)
		}
		return nil
	}
	return fmt.Errorf("standard error carried no `[Tool]` log line naming %q; stderr=%q", tool, sc.stderr)
}

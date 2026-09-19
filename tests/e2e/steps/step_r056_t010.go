package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T010 [BDD-RED] — Then: the run reported the reason "{reason}" for the MCP tool "{tool}" on the server "{server}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the run reported the reason "([^"]*)" for the MCP tool "([^"]*)" on the server "([^"]*)"$`, thenReasonEchoMCP)
	})
}

// thenReasonEchoMCP (必查 呈現結果): the captured stderr carries a
// `[HH:MM:SS] [Tool Reason] {reason}` line whose next non-blank line is the
// NAMESPACED MCP call's `[HH:MM:SS] [Tool Action] mcp_{server}_{tool}(…)` line
// (round 056 / ADR 0025 — the reason row reaches MCP calls).
func thenReasonEchoMCP(ctx context.Context, reason, tool, server string) error {
	sc := scenarioFrom(ctx)
	namespaced := "mcp_" + server + "_" + tool
	lines := strings.Split(sc.stderr, "\n")
	for i, line := range lines {
		if !reasonRowMatches(line, reason) {
			continue
		}
		for j := i + 1; j < len(lines); j++ {
			if strings.TrimSpace(lines[j]) == "" {
				continue
			}
			if actionRowMatches(lines[j], namespaced) {
				return nil
			}
			break
		}
	}
	return fmt.Errorf("standard error carried no `[Tool Reason] %s` line before the action for %q; stderr=%q", reason, namespaced, sc.stderr)
}

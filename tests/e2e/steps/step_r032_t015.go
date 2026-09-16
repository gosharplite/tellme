package steps

import (
	"context"

	"github.com/cucumber/godog"

	mcptest "github.com/gosharplite/tellme/internal/infrastructure/mcp/mcptest"
)

// T015 [BDD-RED] — Given: a remote MCP server "{server}" whose tool "{tool}" reports a tool error
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a remote MCP server "([^"]*)" whose tool "([^"]*)" reports a tool error$`, givenMCPServerToolError)
	})
}

// givenMCPServerToolError arranges a fake MCP server whose {tool} returns an
// MCP-level tool error (isError: true), recorded as MCP_SERVERS.{server}
// (怎麼做 / 權威狀態落地 / 回寫).
func givenMCPServerToolError(ctx context.Context, server, tool string) error {
	sc := scenarioFrom(ctx)
	fake := sc.startMCPFake(server, mcptest.Options{Tool: tool, ToolError: true})
	sc.addMCPServer(server, mcpServerEntry{URL: fake.URL()})
	return nil
}

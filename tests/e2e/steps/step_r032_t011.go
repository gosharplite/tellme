package steps

import (
	"context"

	"github.com/cucumber/godog"

	mcptest "github.com/gosharplite/tellme/internal/infrastructure/mcp/mcptest"
)

// T011 [BDD-RED] — Given: a remote MCP server "{server}" that is marked off
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a remote MCP server "([^"]*)" that is marked off$`, givenMCPServerMarkedOff)
	})
}

// givenMCPServerMarkedOff arranges a recording fake MCP server with
// ENABLED: false (the run must never contact it), recorded as
// MCP_SERVERS.{server} (怎麼做 / 權威狀態落地 / 回寫).
func givenMCPServerMarkedOff(ctx context.Context, server string) error {
	sc := scenarioFrom(ctx)
	disabled := false
	fake := sc.startMCPFake(server, mcptest.Options{})
	sc.addMCPServer(server, mcpServerEntry{URL: fake.URL(), Enabled: &disabled})
	return nil
}

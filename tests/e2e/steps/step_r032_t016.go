package steps

import (
	"context"

	"github.com/cucumber/godog"

	mcptest "github.com/gosharplite/tellme/internal/infrastructure/mcp/mcptest"
)

// T016 [BDD-RED] — Given: a remote MCP server "{server}" that fails the tool call
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a remote MCP server "([^"]*)" that fails the tool call$`, givenMCPServerFailsToolCall)
	})
}

// givenMCPServerFailsToolCall arranges a fake MCP server that advertises the
// `lookup_price` tool (so it is offered) but fails a tools/call at the transport
// layer (HTTP 500), recorded as MCP_SERVERS.{server}. The tool name matches the
// scenario's provider Given (which asks for `mcp_{server}_lookup_price`).
func givenMCPServerFailsToolCall(ctx context.Context, server string) error {
	sc := scenarioFrom(ctx)
	fake := sc.startMCPFake(server, mcptest.Options{Tool: "lookup_price", TransportFail: true})
	sc.addMCPServer(server, mcpServerEntry{URL: fake.URL()})
	return nil
}

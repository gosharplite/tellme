package steps

import (
	"context"

	"github.com/cucumber/godog"

	mcptest "github.com/gosharplite/tellme/internal/infrastructure/mcp/mcptest"
)

// T014 [BDD-RED] — Given: a remote MCP server "{server}" that advertises a tool "{tool}" with a malformed schema
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a remote MCP server "([^"]*)" that advertises a tool "([^"]*)" with a malformed schema$`, givenMCPServerMalformedSchema)
	})
}

// givenMCPServerMalformedSchema arranges a fake MCP server advertising {tool}
// with an UNSAFE schema (a `required` entry with no matching `properties`
// declaration — the #64-mirror shape), recorded as MCP_SERVERS.{server}
// (怎麼做 / 權威狀態落地 / 回寫).
func givenMCPServerMalformedSchema(ctx context.Context, server, tool string) error {
	sc := scenarioFrom(ctx)
	fake := sc.startMCPFake(server, mcptest.Options{Tool: tool, Schema: mcptest.MalformedSchema()})
	sc.addMCPServer(server, mcpServerEntry{URL: fake.URL()})
	return nil
}

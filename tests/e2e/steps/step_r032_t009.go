package steps

import (
	"context"

	"github.com/cucumber/godog"

	mcptest "github.com/gosharplite/tellme/internal/infrastructure/mcp/mcptest"
)

// T009 [BDD-RED] — Given: a remote MCP server "{server}" that offers a tool "{tool}" answering "{result}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a remote MCP server "([^"]*)" that offers a tool "([^"]*)" answering "([^"]*)"$`, givenMCPServerOffersTool)
	})
}

// givenMCPServerOffersTool arranges a reachable fake MCP server offering {tool}
// (answering {result}) and records it as MCP_SERVERS.{server} (怎麼做 /
// 權威狀態落地 / 回寫).
func givenMCPServerOffersTool(ctx context.Context, server, tool, result string) error {
	sc := scenarioFrom(ctx)
	fake := sc.startMCPFake(server, mcptest.Options{Tool: tool, Result: unescapeText(result)})
	sc.addMCPServer(server, mcpServerEntry{URL: fake.URL()})
	return nil
}

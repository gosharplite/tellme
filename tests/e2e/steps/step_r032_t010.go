package steps

import (
	"context"

	"github.com/cucumber/godog"

	mcptest "github.com/gosharplite/tellme/internal/infrastructure/mcp/mcptest"
)

// T010 [BDD-RED] — Given: a remote MCP server "{server}" that never answers
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a remote MCP server "([^"]*)" that never answers$`, givenMCPServerNeverAnswers)
	})
}

// givenMCPServerNeverAnswers arranges a fake that accepts a connection but never
// responds (it holds each request past the fixed fast-fail bound), recorded as
// MCP_SERVERS.{server} (怎麼做 / 權威狀態落地 / 回寫).
func givenMCPServerNeverAnswers(ctx context.Context, server string) error {
	sc := scenarioFrom(ctx)
	fake := sc.startMCPFake(server, mcptest.Options{NeverAnswer: true})
	sc.addMCPServer(server, mcpServerEntry{URL: fake.URL()})
	return nil
}

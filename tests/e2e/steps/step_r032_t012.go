package steps

import (
	"context"

	"github.com/cucumber/godog"

	mcptest "github.com/gosharplite/tellme/internal/infrastructure/mcp/mcptest"
)

// T012 [BDD-RED] — Given: a remote MCP server "{server}" that requires the token "{token}" and offers a tool "{tool}" answering "{result}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a remote MCP server "([^"]*)" that requires the token "([^"]*)" and offers a tool "([^"]*)" answering "([^"]*)"$`, givenMCPServerRequiresToken)
	})
}

// givenMCPServerRequiresToken arranges a fake MCP server that rejects requests
// without `Authorization: Bearer {token}` and records the token as the server's
// TOKEN (怎麼做 / 權威狀態落地 / 回寫).
func givenMCPServerRequiresToken(ctx context.Context, server, token, tool, result string) error {
	sc := scenarioFrom(ctx)
	fake := sc.startMCPFake(server, mcptest.Options{
		Tool:          tool,
		Result:        unescapeText(result),
		RequiredToken: token,
	})
	sc.addMCPServer(server, mcpServerEntry{URL: fake.URL(), Token: token})
	return nil
}

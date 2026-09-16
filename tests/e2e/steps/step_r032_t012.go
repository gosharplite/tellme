package steps

import (
	"context"
	"strings"

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
//
// The configured TOKEN is routed through a ${VAR} reference (issue #67 /
// round-032 SC-002): the config carries `TOKEN: "${TELLME_E2E_MCP_TOKEN_<server>}"`
// and the environment variable resolves to the token. This is the real-world
// shape (`TOKEN: "${GITHUB_TOKEN}"`), and it makes the scenario a genuine witness
// of MCP_SERVERS expansion — without it the literal string is sent, the server
// 401s, and the "received the token" Then fails.
func givenMCPServerRequiresToken(ctx context.Context, server, token, tool, result string) error {
	sc := scenarioFrom(ctx)
	fake := sc.startMCPFake(server, mcptest.Options{
		Tool:          tool,
		Result:        unescapeText(result),
		RequiredToken: token,
	})
	envName := "TELLME_E2E_MCP_TOKEN_" + strings.ToUpper(server)
	sc.setEnv(envName, token)
	sc.addMCPServer(server, mcpServerEntry{URL: fake.URL(), Token: "${" + envName + "}"})
	return nil
}

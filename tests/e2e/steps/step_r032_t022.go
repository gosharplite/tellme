package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T022 [BDD-RED] — Then: the MCP server "{server}" received the token "{token}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the MCP server "([^"]*)" received the token "([^"]*)"$`, thenMCPReceivedToken)
	})
}

// thenMCPReceivedToken (必查 權威狀態): the fake recorded an
// `Authorization: Bearer <token>` header on its discovery/call requests.
func thenMCPReceivedToken(ctx context.Context, server, token string) error {
	sc := scenarioFrom(ctx)
	fake := sc.mcpFake(server)
	if fake == nil {
		return fmt.Errorf("no fake MCP server registered for %q", server)
	}
	if !fake.ReceivedAuthorization(token) {
		return fmt.Errorf("the MCP server %q never received the bearer token %q", server, token)
	}
	return nil
}

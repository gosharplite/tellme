package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T021 [BDD-RED] — Then: tellme never contacted the MCP server "{server}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme never contacted the MCP server "([^"]*)"$`, thenNeverContactedMCPServer)
	})
}

// thenNeverContactedMCPServer (必查 權威狀態): the recording fake recorded ZERO
// connections — a server marked off is never dialed.
func thenNeverContactedMCPServer(ctx context.Context, server string) error {
	sc := scenarioFrom(ctx)
	fake := sc.mcpFake(server)
	if fake == nil {
		return fmt.Errorf("no fake MCP server registered for %q", server)
	}
	if n := fake.ConnectionCount(); n != 0 {
		return fmt.Errorf("the MCP server %q was contacted %d time(s); want zero", server, n)
	}
	return nil
}

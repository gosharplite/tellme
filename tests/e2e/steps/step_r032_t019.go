package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T019 [BDD-RED] — Then: tellme called the tool "{tool}" on the MCP server "{server}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme called the tool "([^"]*)" on the MCP server "([^"]*)"$`, thenCalledMCPTool)
	})
}

// thenCalledMCPTool (必查 權威狀態): the fake MCP server recorded a call for
// {tool} — the result was produced on the server, not fabricated.
func thenCalledMCPTool(ctx context.Context, tool, server string) error {
	sc := scenarioFrom(ctx)
	fake := sc.mcpFake(server)
	if fake == nil {
		return fmt.Errorf("no fake MCP server registered for %q", server)
	}
	if !fake.Called(tool) {
		return fmt.Errorf("the MCP server %q did not record a call to %q", server, tool)
	}
	return nil
}

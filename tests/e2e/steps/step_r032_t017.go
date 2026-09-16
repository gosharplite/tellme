package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T017 [BDD-RED] — Then: the request offered the tool "{tool}" from the MCP server "{server}" alongside the agent tools
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the request offered the tool "([^"]*)" from the MCP server "([^"]*)" alongside the agent tools$`, thenOfferedMCPToolAlongside)
	})
}

// thenOfferedMCPToolAlongside (必查 權威狀態): the recorded request offered the
// namespaced `mcp_{server}_{tool}` AND every native agent tool — the MCP tools
// augment, never replace, the native set.
func thenOfferedMCPToolAlongside(ctx context.Context, tool, server string) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no provider fake recorded a request")
	}
	names := f.ToolNamesAt(-1)
	want := "mcp_" + server + "_" + tool
	if !containsString(names, want) {
		return fmt.Errorf("the request did not offer %q; offered: %v", want, names)
	}
	for _, native := range nativeAgentTools {
		if !containsString(names, native) {
			return fmt.Errorf("the request dropped the native tool %q; offered: %v", native, names)
		}
	}
	return nil
}

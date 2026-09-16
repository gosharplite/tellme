package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T018 [BDD-RED] — Then: the request offered no tool from the MCP server "{server}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the request offered no tool from the MCP server "([^"]*)"$`, thenOfferedNoMCPTool)
	})
}

// thenOfferedNoMCPTool (必查 權威狀態): no offered tool name begins with
// `mcp_{server}_` — a server marked off, or one whose tools were all skipped,
// contributes none.
func thenOfferedNoMCPTool(ctx context.Context, server string) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no provider fake recorded a request")
	}
	prefix := "mcp_" + server + "_"
	for _, n := range f.ToolNamesAt(-1) {
		if strings.HasPrefix(n, prefix) {
			return fmt.Errorf("the request offered %q from the server %q; want none", n, server)
		}
	}
	return nil
}

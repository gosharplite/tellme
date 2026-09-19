package steps

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cucumber/godog"
)

// T007 [BDD-RED] — Then: the MCP server "{server}" received only the arguments its tool expects
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the MCP server "([^"]*)" received only the arguments its tool expects$`, thenMCPServerReceivedOnlyPayload)
	})
}

// thenMCPServerReceivedOnlyPayload (必查 呈現結果): every tool call the fake MCP
// server recorded carried NO `reason` key and NO `MCP_PAYLOAD` wrapper — tellme
// forwarded only the server's own argument object (round 056 / ADR 0025 D2).
func thenMCPServerReceivedOnlyPayload(ctx context.Context, server string) error {
	sc := scenarioFrom(ctx)
	fake := sc.mcpFake(server)
	if fake == nil {
		return fmt.Errorf("no fake MCP server %q was started", server)
	}
	calls := fake.AllReceivedArguments()
	if len(calls) == 0 {
		return fmt.Errorf("the MCP server %q recorded no tool call", server)
	}
	for _, raw := range calls {
		var m map[string]json.RawMessage
		if err := json.Unmarshal([]byte(raw), &m); err != nil {
			return fmt.Errorf("the MCP server %q recorded unparseable arguments %q", server, raw)
		}
		if _, ok := m["reason"]; ok {
			return fmt.Errorf("the MCP server %q received a `reason` key (tellme must not forward it); args=%q", server, raw)
		}
		if _, ok := m[mcpPayloadKey]; ok {
			return fmt.Errorf("the MCP server %q received the `MCP_PAYLOAD` wrapper (tellme must forward only the payload); args=%q", server, raw)
		}
	}
	return nil
}

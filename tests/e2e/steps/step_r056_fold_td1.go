package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// Round-056 review fold (TD-056-1): the boundary carriers for "a refused call
// never reaches the server". `CallCount` is the force-bearing observable — a
// refused attempt must not increment it.

func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the MCP server "([^"]*)" received no call$`, thenMCPServerReceivedNoCall)
		ctx.Then(`^the MCP server "([^"]*)" received exactly one call$`, thenMCPServerReceivedExactlyOneCall)
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint asks tellme to use the MCP tool "([^"]*)" from the server "([^"]*)" with the reason "([^"]*)" and a stray argument, and then answers with "([^"]*)"$`, givenProviderMCPStrayArgument)
	})
}

// thenMCPServerReceivedNoCall (必查 呈現結果): the fake MCP server recorded no
// tool call — the refused call was stopped before CallTool.
func thenMCPServerReceivedNoCall(ctx context.Context, server string) error {
	sc := scenarioFrom(ctx)
	fake := sc.mcpFake(server)
	if fake == nil {
		return fmt.Errorf("no fake MCP server %q was started", server)
	}
	if n := fake.CallCount(); n != 0 {
		return fmt.Errorf("the MCP server %q received %d call(s); a refused call must never reach it", server, n)
	}
	return nil
}

// thenMCPServerReceivedExactlyOneCall (必查 呈現結果): exactly one tool call
// reached the server — the refused attempt did not, the conforming retry did.
func thenMCPServerReceivedExactlyOneCall(ctx context.Context, server string) error {
	sc := scenarioFrom(ctx)
	fake := sc.mcpFake(server)
	if fake == nil {
		return fmt.Errorf("no fake MCP server %q was started", server)
	}
	if n := fake.CallCount(); n != 1 {
		return fmt.Errorf("the MCP server %q received %d call(s), want exactly 1", server, n)
	}
	return nil
}

// givenProviderMCPStrayArgument scripts a single MCP call carrying a top-level
// key other than `reason`/`MCP_PAYLOAD` — a shape violation the adapter refuses
// before contacting the server (then the answer).
func givenProviderMCPStrayArgument(ctx context.Context, provider, tool, server, reason, answer string) error {
	sc := scenarioFrom(ctx)
	env := mcpEnvelopeWithStray(unescapeText(reason))
	name := "mcp_" + server + "_" + tool
	f := sc.newFake()
	f.Script(
		fakeprovider.Reply{ToolName: name, Arguments: env},
		fakeprovider.Reply{Answer: unescapeText(answer)},
	)
	sc.scriptedTool = name
	sc.scriptedAnswer = unescapeText(answer)
	sc.scriptedAnswerSet = true
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

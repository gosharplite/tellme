package steps

import (
	"context"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// T004 [BDD-ALIGN] — Given: a configured provider "{provider}" whose endpoint asks tellme to use the MCP tool "{tool}" from the server "{server}" with the reason "{reason}" and then answers with "{answer}"
//
// Round 056 (ADR 0025): the MCP call now carries tellme's envelope — a top-level
// `reason` (rendered by tellme) plus `MCP_PAYLOAD` (the server's own arguments).
// This replaces the retired no-reason sentence (the universal gate would refuse a
// reason-less MCP call).
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint asks tellme to use the MCP tool "([^"]*)" from the server "([^"]*)" with the reason "([^"]*)" and then answers with "([^"]*)"$`, givenProviderAsksMCPToolWithReason)
	})
}

// givenProviderAsksMCPToolWithReason scripts the fake provider to return a tool
// call naming the NAMESPACED MCP tool `mcp_{server}_{tool}` carrying the
// round-056 envelope `{"reason":"{reason}","MCP_PAYLOAD":{}}`, then a final answer
// {answer}; and writes the resolvable default configuration (怎麼做 / 權威狀態落地 / 回寫).
func givenProviderAsksMCPToolWithReason(ctx context.Context, provider, tool, server, reason, answer string) error {
	sc := scenarioFrom(ctx)
	env := mcpEnvelopeJSON(unescapeText(reason), "{}")
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

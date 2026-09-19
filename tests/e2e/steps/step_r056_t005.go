package steps

import (
	"context"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// T005 [BDD-RED] — Given: a configured provider "{provider}" whose endpoint first asks tellme to use the MCP tool "{tool}" from the server "{server}" without a reason and then with the reason "{reason}" and then answers with "{answer}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint first asks tellme to use the MCP tool "([^"]*)" from the server "([^"]*)" without a reason and then with the reason "([^"]*)" and then answers with "([^"]*)"$`, givenProviderMCPFirstReasonlessThenReason)
	})
}

// givenProviderMCPFirstReasonlessThenReason scripts a three-step exchange: a
// reason-less MCP call (a bare object — REFUSED by the universal gate, the server
// is never contacted), then the same call carrying the round-056 envelope, then
// the answer (round 056 / ADR 0025 D2/D3).
func givenProviderMCPFirstReasonlessThenReason(ctx context.Context, provider, tool, server, reason, answer string) error {
	sc := scenarioFrom(ctx)
	name := "mcp_" + server + "_" + tool
	f := sc.newFake()
	f.Script(
		fakeprovider.Reply{ToolName: name, Arguments: "{}"},
		fakeprovider.Reply{ToolName: name, Arguments: mcpEnvelopeJSON(unescapeText(reason), "{}")},
		fakeprovider.Reply{Answer: unescapeText(answer)},
	)
	sc.scriptedTool = name
	sc.scriptedAnswer = unescapeText(answer)
	sc.scriptedAnswerSet = true
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

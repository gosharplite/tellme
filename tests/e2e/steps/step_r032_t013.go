package steps

import (
	"context"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// T013 [BDD-RED] — Given: a configured provider "{provider}" whose endpoint asks tellme to use the MCP tool "{tool}" from the server "{server}" and then answers with "{answer}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint asks tellme to use the MCP tool "([^"]*)" from the server "([^"]*)" and then answers with "([^"]*)"$`, givenProviderAsksMCPTool)
	})
}

// givenProviderAsksMCPTool scripts the fake provider to return a tool call
// naming the NAMESPACED MCP tool `mcp_{server}_{tool}` (arguments `{}`), then a
// final answer {answer}; and writes the resolvable default configuration
// (怎麼做 / 權威狀態落地 / 回寫).
func givenProviderAsksMCPTool(ctx context.Context, provider, tool, server, answer string) error {
	sc := scenarioFrom(ctx)
	answer = unescapeText(answer)
	name := "mcp_" + server + "_" + tool
	f := sc.newFake()
	f.Script(
		fakeprovider.Reply{ToolName: name, Arguments: "{}"},
		fakeprovider.Reply{Answer: answer},
	)
	sc.scriptedTool = name
	sc.scriptedAnswer = answer
	sc.scriptedAnswerSet = true
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

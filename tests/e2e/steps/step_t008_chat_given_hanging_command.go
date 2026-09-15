package steps

import (
	"context"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// T008 [BDD-RED] — Given: a configured provider "{provider}" whose endpoint runs a command that never returns within a short limit and then answers with "{answer}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint runs a command that never returns within a short limit and then answers with "([^"]*)"$`, givenProviderHangingCommand)
	})
}

// givenProviderHangingCommand (怎麼做 / 權威狀態落地 / 回寫): write the resolvable
// default config selecting {provider}; script the fake to return an
// `execute_command` tool call carrying a short `timeout` and a command that never
// returns (stopped at the per-call deadline), then a final answer {answer}.
func givenProviderHangingCommand(ctx context.Context, provider, answer string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	f.Script(
		fakeprovider.Reply{ToolName: "execute_command", Arguments: `{"command":"sleep 60","timeout":1,"reason":"run a hanging command"}`},
		fakeprovider.Reply{Answer: unescapeText(answer)},
	)
	sc.scriptedAnswer = unescapeText(answer)
	sc.scriptedAnswerSet = true
	sc.scriptedTool = "execute_command"
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

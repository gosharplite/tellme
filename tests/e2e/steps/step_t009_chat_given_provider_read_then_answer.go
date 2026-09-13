package steps

import (
	"context"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// T009 — Given: a configured provider "{provider}" whose endpoint asks tellme to read "{path}" and then answers with "{answer}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint asks tellme to read "([^"]*)" and then answers with "([^"]*)"$`, givenProviderReadThenAnswer)
	})
}

// givenProviderReadThenAnswer (怎麼做 / 權威狀態落地 / 回寫): write the resolvable
// default config selecting {provider}; script the fake to return a `read_files`
// tool-call for {path} on the first request, then a final answer {answer}.
func givenProviderReadThenAnswer(ctx context.Context, provider, path, answer string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	f.Script(
		fakeprovider.Reply{ToolName: "read_files", Arguments: readArgs(path)},
		fakeprovider.Reply{Answer: unescapeText(answer)},
	)
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

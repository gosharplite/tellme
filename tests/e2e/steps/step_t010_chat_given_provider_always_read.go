package steps

import (
	"context"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// T010 — Given: a configured provider "{provider}" whose endpoint always asks tellme to read "{path}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint always asks tellme to read "([^"]*)"$`, givenProviderAlwaysRead)
	})
}

// givenProviderAlwaysRead (怎麼做 / 權威狀態落地 / 回寫): write the config selecting
// {provider}; script the fake to return the same `read_files` tool-call on every
// request (a single-element script repeats), so the loop can never reach a final
// answer and must hit its bound.
func givenProviderAlwaysRead(ctx context.Context, provider, path string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	f.Script(fakeprovider.Reply{ToolName: "read_files", Arguments: readArgs(path)})
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

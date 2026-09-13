package steps

import (
	"context"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// T011 — Given: a configured provider "{provider}" whose endpoint asks for a tool that is not available
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint asks for a tool that is not available$`, givenProviderUnknownTool)
	})
}

// givenProviderUnknownTool (怎麼做 / 權威狀態落地 / 回寫): write the config selecting
// {provider}; script the fake to return a tool-call naming a tool tellme does
// not provide, so the loop fails fast with the tool class phrase.
func givenProviderUnknownTool(ctx context.Context, provider string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	f.Script(fakeprovider.Reply{ToolName: "time_travel", Arguments: "{}"})
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

package steps

import (
	"context"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// T006 [BDD-RED] — Given: a configured provider "{provider}" whose reply is cut off at the output limit before finishing its answer
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose reply is cut off at the output limit before finishing its answer$`, givenProviderCutoffAnswer)
	})
}

// givenProviderCutoffAnswer (怎麼做 / 權威狀態落地 / 回寫): write the configuration and
// script the fake to return ONE response whose OpenAI-compatible `finish_reason`
// is "length" and whose message carries only PARTIAL answer text (no tool call) —
// an answer cut off before it finished.
func givenProviderCutoffAnswer(ctx context.Context, provider string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	f.Script(fakeprovider.Reply{Answer: "The answer was cut off", FinishReason: "length"})
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

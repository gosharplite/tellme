package steps

import (
	"context"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// T007 [BDD-RED] — Given: a configured Gemini provider "{provider}" whose reply is cut off at the output limit before finishing its answer
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured Gemini provider "([^"]*)" whose reply is cut off at the output limit before finishing its answer$`, givenGeminiCutoffAnswer)
	})
}

// givenGeminiCutoffAnswer (怎麼做 / 權威狀態落地 / 回寫): arrange a Vertex-shaped
// `gemini` provider (service-account key file + config), and script the fake to
// return ONE Vertex `:generateContent` response whose
// `candidates[0].finishReason` is "MAX_TOKENS" carrying only partial text (no
// `functionCall`) — a Gemini answer cut off at the output cap.
func givenGeminiCutoffAnswer(ctx context.Context, provider string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	f.VertexMode()
	f.Script(fakeprovider.Reply{Answer: "The answer was cut off", FinishReason: "MAX_TOKENS"})
	sc.registerFake(provider, f)
	keyPath, err := sc.writeServiceAccountKey("secrets/key.json", f.URL()+"/token")
	if err != nil {
		return err
	}
	return sc.writeGeminiConfig(provider, "gemini-3.8-flash", f.URL(), keyPath, 40960)
}

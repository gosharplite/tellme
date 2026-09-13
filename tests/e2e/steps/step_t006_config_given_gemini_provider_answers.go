package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T006 — Given: a configured Gemini provider "{provider}" whose endpoint answers with "{answer}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured Gemini provider "([^"]*)" whose endpoint answers with "([^"]*)"$`, givenGeminiProviderAnswers)
	})
}

// givenGeminiProviderAnswers (構造 / 權威狀態落地): starts a Vertex-mode fake,
// writes a service-account key file whose token_uri is the fake's token endpoint,
// and writes the default configuration selecting a `gemini` provider whose
// Vertex-shaped URL points at the fake (怎麼做 / 回寫).
func givenGeminiProviderAnswers(ctx context.Context, provider, answer string) error {
	sc := scenarioFrom(ctx)
	answer = unescapeText(answer)
	f := sc.newFake()
	f.VertexMode()
	f.Answer(answer)
	sc.scriptedAnswer = answer
	sc.scriptedAnswerSet = true
	sc.registerFake(provider, f)
	keyPath, err := sc.writeServiceAccountKey("secrets/key.json", f.URL()+"/token")
	if err != nil {
		return err
	}
	return sc.writeGeminiConfig(provider, "gemini-3.8-flash", f.URL(), keyPath, 40960)
}

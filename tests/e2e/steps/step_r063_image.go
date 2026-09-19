package steps

import (
	"context"
	"encoding/json"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// Round 063 (ADR 0033) E2E steps: the Gemini-family image journey's Given
// sentences. The Then sentences are the round-062 ones (widened family-aware),
// so this file only arranges a Vertex-shaped provider that records its request.
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured Gemini provider "([^"]*)" that can take images and whose endpoint asks tellme to read "([^"]*)" and then answers with "([^"]*)"$`, givenGeminiVisionProviderReads)
		ctx.Given(`^a configured Gemini provider "([^"]*)" that can take images whose endpoint answers with "([^"]*)"$`, givenGeminiVisionProviderAnswers)
		ctx.Given(`^a configured Gemini provider "([^"]*)" that cannot take images whose endpoint answers with "([^"]*)"$`, givenGeminiBlindProviderAnswers)
	})
}

// givenGeminiVisionProviderReads arranges a Vertex-shaped gemini provider whose
// fake scripts one `read_image` call (for {path}), then answers {answer}, with
// the provider entry declaring `VISION: true`.
func givenGeminiVisionProviderReads(ctx context.Context, provider, path, answer string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	f.VertexMode()
	args, err := json.Marshal(map[string]any{
		"filepath": path,
		"reason":   "look at the picture",
	})
	if err != nil {
		return err
	}
	f.Script(
		fakeprovider.Reply{ToolName: "read_image", Arguments: string(args)},
		fakeprovider.Reply{Answer: unescapeText(answer)},
	)
	sc.scriptedAnswer = unescapeText(answer)
	sc.scriptedAnswerSet = true
	sc.scriptedTool = "read_image"
	sc.registerFake(provider, f)
	keyPath, err := sc.writeServiceAccountKey("secrets/key.json", f.URL()+"/token")
	if err != nil {
		return err
	}
	if err := sc.writeGeminiConfig(provider, "gemini-3.8-flash", f.URL(), keyPath, 40960); err != nil {
		return err
	}
	return sc.setSelectedProviderVision(true)
}

// givenGeminiVisionProviderAnswers arranges a Vertex-shaped gemini provider that
// declares `VISION: true` and just answers {answer} (no tool call).
func givenGeminiVisionProviderAnswers(ctx context.Context, provider, answer string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	f.VertexMode()
	f.Answer(unescapeText(answer))
	sc.scriptedAnswer = unescapeText(answer)
	sc.scriptedAnswerSet = true
	sc.registerFake(provider, f)
	keyPath, err := sc.writeServiceAccountKey("secrets/key.json", f.URL()+"/token")
	if err != nil {
		return err
	}
	if err := sc.writeGeminiConfig(provider, "gemini-3.8-flash", f.URL(), keyPath, 40960); err != nil {
		return err
	}
	return sc.setSelectedProviderVision(true)
}

// givenGeminiBlindProviderAnswers arranges a Vertex-shaped gemini provider
// WITHOUT vision (no `VISION` key) that just answers {answer}.
func givenGeminiBlindProviderAnswers(ctx context.Context, provider, answer string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	f.VertexMode()
	f.Answer(unescapeText(answer))
	sc.scriptedAnswer = unescapeText(answer)
	sc.scriptedAnswerSet = true
	sc.registerFake(provider, f)
	keyPath, err := sc.writeServiceAccountKey("secrets/key.json", f.URL()+"/token")
	if err != nil {
		return err
	}
	if err := sc.writeGeminiConfig(provider, "gemini-3.8-flash", f.URL(), keyPath, 40960); err != nil {
		return err
	}
	return sc.setSelectedProviderVision(false)
}

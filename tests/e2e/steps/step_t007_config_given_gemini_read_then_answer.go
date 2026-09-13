package steps

import (
	"context"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// T007 — Given: a configured Gemini provider "{provider}" whose endpoint asks tellme to read "{path}" and then answers with "{answer}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured Gemini provider "([^"]*)" whose endpoint asks tellme to read "([^"]*)" and then answers with "([^"]*)"$`, givenGeminiProviderReadThenAnswer)
	})
}

// givenGeminiProviderReadThenAnswer scripts the Vertex-mode fake for a two-step
// exchange — a `read_files` functionCall for {path}, then a Vertex answer
// {answer} — and arranges the gemini provider + service-account credential.
func givenGeminiProviderReadThenAnswer(ctx context.Context, provider, path, answer string) error {
	sc := scenarioFrom(ctx)
	answer = unescapeText(answer)
	f := sc.newFake()
	f.VertexMode()
	f.Script(
		fakeprovider.Reply{ToolName: "read_files", Arguments: readArgs(path)},
		fakeprovider.Reply{Answer: answer},
	)
	sc.scriptedAnswer = answer
	sc.scriptedAnswerSet = true
	sc.registerFake(provider, f)
	keyPath, err := sc.writeServiceAccountKey("secrets/key.json", f.URL()+"/token")
	if err != nil {
		return err
	}
	return sc.writeGeminiConfig(provider, "gemini-3.8-flash", f.URL(), keyPath, 40960)
}

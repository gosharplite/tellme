package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// Round 072 (ADR 0044; closes #149) — the Gemini/Vertex usage decode E2E steps.
// Kept in ONE file (the Zero-Shared-Edits split exists for parallel dispatch).
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured Gemini provider "([^"]*)" whose endpoint answers with "([^"]*)" and reports the token usage:$`, givenGeminiProviderReportsTokenUsage)
	})
}

// givenGeminiProviderReportsTokenUsage (怎麼做 / 權威狀態落地 / 回寫): starts a
// Vertex-mode fake, writes a service-account key + a `gemini` config pointing at
// the fake, and scripts the fake to answer {answer} with a DETAILED Vertex
// `usageMetadata` (prompt/candidates/total + cachedContentTokenCount +
// thoughtsTokenCount). The Vertex family is DISJOINT, so the wire
// `candidatesTokenCount` is the plain completion (not +thinking).
func givenGeminiProviderReportsTokenUsage(ctx context.Context, provider, answer string, table *godog.Table) error {
	sc := scenarioFrom(ctx)
	answer = unescapeText(answer)
	prompt, cached, completion, thinking, err := usageFromTable(table)
	if err != nil {
		return err
	}
	f := sc.newFake()
	f.VertexMode()
	f.Answer(answer)
	f.ReportUsageDetails(prompt, cached, completion, thinking)
	sc.scriptedAnswer = answer
	sc.scriptedAnswerSet = true
	sc.registerFake(provider, f)
	keyPath, err := sc.writeServiceAccountKey("secrets/key.json", f.URL()+"/token")
	if err != nil {
		return err
	}
	return sc.writeGeminiConfig(provider, "gemini-3.8-flash", f.URL(), keyPath, 40960)
}

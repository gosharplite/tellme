package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T008 — Given: a configured Gemini provider "{provider}" whose key file is missing
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured Gemini provider "([^"]*)" whose key file is missing$`, givenGeminiProviderKeyMissing)
	})
}

// givenGeminiProviderKeyMissing arranges a gemini provider whose API_KEY points
// at a `.json` path that does not exist (構造 / 權威狀態落地: the credential file
// is missing).
func givenGeminiProviderKeyMissing(ctx context.Context, provider string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	f.VertexMode()
	sc.registerFake(provider, f)
	missing := sc.homePath("secrets/key.json") // never written
	return sc.writeGeminiConfig(provider, "gemini-3.8-flash", f.URL(), missing, 40960)
}

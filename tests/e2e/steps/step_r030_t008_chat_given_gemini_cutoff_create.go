package steps

import (
	"context"
	"encoding/json"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// T008 [BDD-RED] — Given: a configured Gemini provider "{provider}" whose reply is cut off at the output limit while creating the file "{path}" with the content "{content}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured Gemini provider "([^"]*)" whose reply is cut off at the output limit while creating the file "([^"]*)" with the content "([^"]*)"$`, givenGeminiCutoffCreate)
	})
}

// givenGeminiCutoffCreate (怎麼做 / 權威狀態落地 / 回寫): arrange a Vertex-shaped
// `gemini` provider (service-account key file + config), and script the fake to
// return ONE Vertex response whose `candidates[0].finishReason` is "MAX_TOKENS"
// carrying a `functionCall` for `write_file` (args filepath = {path}, content =
// {content}, a reason) — the Gemini function-call truncation site (B1).
func givenGeminiCutoffCreate(ctx context.Context, provider, path, content string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	f.VertexMode()
	args, err := json.Marshal(map[string]any{
		"filepath": path,
		"content":  unescapeText(content),
		"reason":   "create the file",
	})
	if err != nil {
		return err
	}
	f.Script(fakeprovider.Reply{ToolName: "write_file", Arguments: string(args), FinishReason: "MAX_TOKENS"})
	sc.scriptedTool = "write_file"
	sc.registerFake(provider, f)
	keyPath, err := sc.writeServiceAccountKey("secrets/key.json", f.URL()+"/token")
	if err != nil {
		return err
	}
	return sc.writeGeminiConfig(provider, "gemini-3.8-flash", f.URL(), keyPath, 40960)
}

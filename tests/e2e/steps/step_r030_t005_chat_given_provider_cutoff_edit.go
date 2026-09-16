package steps

import (
	"context"
	"encoding/json"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// T005 [BDD-RED] — Given: a configured provider "{provider}" whose reply is cut off at the output limit while editing the file "{path}" replacing "{old}" with "{new}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose reply is cut off at the output limit while editing the file "([^"]*)" replacing "([^"]*)" with "([^"]*)"$`, givenProviderCutoffEdit)
	})
}

// givenProviderCutoffEdit (怎麼做 / 權威狀態落地 / 回寫): write the configuration and
// script the fake to return ONE response whose OpenAI-compatible `finish_reason`
// is "length" and whose message carries a `replace_text` tool call (arguments
// filepath = {path}, old_text = {old}, new_text = {new}, a reason) — a reply cut
// off mid-tool-call, with no final answer.
func givenProviderCutoffEdit(ctx context.Context, provider, path, oldText, newText string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	args, err := json.Marshal(map[string]any{
		"filepath": path,
		"old_text": unescapeText(oldText),
		"new_text": unescapeText(newText),
		"reason":   "edit the file",
	})
	if err != nil {
		return err
	}
	f.Script(fakeprovider.Reply{ToolName: "replace_text", Arguments: string(args), FinishReason: "length"})
	sc.scriptedTool = "replace_text"
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

package steps

import (
	"context"
	"encoding/json"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// T004 [BDD-RED] — Given: a configured provider "{provider}" whose reply is cut off at the output limit while creating the file "{path}" with the content "{content}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose reply is cut off at the output limit while creating the file "([^"]*)" with the content "([^"]*)"$`, givenProviderCutoffCreate)
	})
}

// givenProviderCutoffCreate (怎麼做 / 權威狀態落地 / 回寫): write the configuration and
// script the fake to return ONE response whose OpenAI-compatible `finish_reason`
// is "length" and whose message carries a `write_file` tool call (arguments
// filepath = {path}, content = {content}, a reason) — a reply cut off
// mid-tool-call, with no final answer.
func givenProviderCutoffCreate(ctx context.Context, provider, path, content string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	args, err := json.Marshal(map[string]any{
		"filepath": path,
		"content":  unescapeText(content),
		"reason":   "create the file",
	})
	if err != nil {
		return err
	}
	f.Script(fakeprovider.Reply{ToolName: "write_file", Arguments: string(args), FinishReason: "length"})
	sc.scriptedTool = "write_file"
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

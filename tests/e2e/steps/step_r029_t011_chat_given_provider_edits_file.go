package steps

import (
	"context"
	"encoding/json"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// T011 [BDD-RED] — Given: a configured provider "{provider}" whose endpoint edits the file "{path}" replacing "{old}" with "{new}" and then answers with "{answer}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint edits the file "([^"]*)" replacing "([^"]*)" with "([^"]*)" and then answers with "([^"]*)"$`, givenProviderEditsFile)
	})
}

// givenProviderEditsFile (怎麼做 / 權威狀態落地 / 回寫): write the configuration and
// script the fake to return a `replace_text` tool-call response (arguments
// filepath = {path}, old_text = {old}, new_text = {new}, a reason), then a final
// answer {answer}.
func givenProviderEditsFile(ctx context.Context, provider, path, oldText, newText, answer string) error {
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
	f.Script(
		fakeprovider.Reply{ToolName: "replace_text", Arguments: string(args)},
		fakeprovider.Reply{Answer: unescapeText(answer)},
	)
	sc.scriptedAnswer = unescapeText(answer)
	sc.scriptedAnswerSet = true
	sc.scriptedTool = "replace_text"
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

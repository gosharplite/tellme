package steps

import (
	"context"
	"encoding/json"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// T010 [BDD-RED] — Given: a configured provider "{provider}" whose endpoint creates the file "{path}" with the content "{content}" and then answers with "{answer}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint creates the file "([^"]*)" with the content "([^"]*)" and then answers with "([^"]*)"$`, givenProviderCreatesFile)
	})
}

// givenProviderCreatesFile (怎麼做 / 權威狀態落地 / 回寫): write the configuration and
// script the fake to return a `write_file` tool-call response (arguments
// filepath = {path}, content = {content}, a reason), then a final answer {answer}.
func givenProviderCreatesFile(ctx context.Context, provider, path, content, answer string) error {
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
	f.Script(
		fakeprovider.Reply{ToolName: "write_file", Arguments: string(args)},
		fakeprovider.Reply{Answer: unescapeText(answer)},
	)
	sc.scriptedAnswer = unescapeText(answer)
	sc.scriptedAnswerSet = true
	sc.scriptedTool = "write_file"
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

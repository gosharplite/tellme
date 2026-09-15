package steps

import (
	"context"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// T012 [BDD-RED] — Given: a configured provider "{provider}" whose endpoint asks tellme to read more files than one request allows and then answers with "{answer}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint asks tellme to read more files than one request allows and then answers with "([^"]*)"$`, givenProviderReadTooMany)
	})
}

// givenProviderReadTooMany (怎麼做 / 權威狀態落地 / 回寫): script the fake to return
// a read_files tool call listing more than 50 file paths, then the answer
// {answer}; the tool must report the too-many-files error inside its result.
func givenProviderReadTooMany(ctx context.Context, provider, answer string) error {
	sc := scenarioFrom(ctx)
	paths := make([]string, 0, 51)
	for i := 0; i < 51; i++ {
		paths = append(paths, r021TooManyFileName(i))
	}
	f := sc.newFake()
	f.Script(
		fakeprovider.Reply{ToolName: "read_files", Arguments: readFilesArgs(paths)},
		fakeprovider.Reply{Answer: unescapeText(answer)},
	)
	sc.scriptedAnswer = unescapeText(answer)
	sc.scriptedAnswerSet = true
	sc.scriptedTool = "read_files"
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

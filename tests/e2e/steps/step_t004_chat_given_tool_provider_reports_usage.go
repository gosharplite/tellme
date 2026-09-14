package steps

import (
	"context"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// T004 — Given: a configured provider "{provider}" whose endpoint asks tellme to read "{path}" and then answers with "{answer}" and reports the token usage:
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint asks tellme to read "([^"]*)" and then answers with "([^"]*)" and reports the token usage:$`, givenToolProviderReportsTokenUsage)
	})
}

// givenToolProviderReportsTokenUsage (怎麼做 / 權威狀態落地 / 回寫): script the fake
// to return a `read_files` tool-call for {path} carrying the JSON `usage` block,
// then a final answer {answer} carrying the same block — so BOTH provider calls
// report usage and the turn accumulates two usage-bearing calls ($#1 < $#2).
func givenToolProviderReportsTokenUsage(ctx context.Context, provider, path, answer string, table *godog.Table) error {
	sc := scenarioFrom(ctx)
	answer = unescapeText(answer)
	prompt, cached, completion, thinking, err := usageFromTable(table)
	if err != nil {
		return err
	}
	f := sc.newFake()
	f.ReportUsageDetails(prompt, cached, completion, thinking)
	f.Script(
		fakeprovider.Reply{ToolName: "read_files", Arguments: readArgs(path)},
		fakeprovider.Reply{Answer: answer},
	)
	sc.scriptedAnswer = answer
	sc.scriptedAnswerSet = true
	sc.scriptedTool = "read_files"
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

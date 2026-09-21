package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// Round 076 (issue #154; ADR 0048) — the unknown-tool-name fold-back.

func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint first asks for a tool that is not available and then asks tellme to read "([^"]*)" and then answers with "([^"]*)"$`, givenProviderUnknownThenReadThenAnswer)
		ctx.Then(`^the run reported the unavailable tool "([^"]*)"$`, thenUnknownToolReported)
	})
}

// givenProviderUnknownThenReadThenAnswer (怎麼做 / 權威狀態落地 / 回寫): script a
// three-step exchange — an UNKNOWN tool call (tellme folds back a recoverable
// result and CONTINUES), then a `read_files` call for {path}, then the answer
// {answer} — so a one-off unknown name is recovered and the turn finishes (round
// 076).
func givenProviderUnknownThenReadThenAnswer(ctx context.Context, provider, path, answer string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	f.Script(
		fakeprovider.Reply{ToolName: "time_travel", Arguments: "{}"},
		fakeprovider.Reply{ToolName: "read_files", Arguments: readArgs(path)},
		fakeprovider.Reply{Answer: unescapeText(answer)},
	)
	sc.scriptedTool = "read_files"
	sc.scriptedAnswer = unescapeText(answer)
	sc.scriptedAnswerSet = true
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

// thenUnknownToolReported (必查 權威狀態): the loop fed back a recoverable result
// naming the unknown tool — a `tool`-role message carrying `no tool named "{tool}"`
// (round 076). A terminal abort (the `the tool request failed` phrase) would have
// no such fold-back.
func thenUnknownToolReported(ctx context.Context, tool string) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	want := fmt.Sprintf("no tool named %q", tool)
	for _, msgs := range toolRounds(f) {
		for _, m := range msgs {
			if m.Role == "tool" && strings.Contains(m.Content, want) {
				return nil
			}
		}
	}
	return fmt.Errorf("no recoverable result naming the unknown tool %q was fed back", tool)
}

package steps

import (
	"context"
	"encoding/json"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// T006 [BDD-RED] — Given: a configured provider "{provider}" whose endpoint first asks tellme to create the file "{path}" without a reason, then with the content "{content}" and the reason "{reason}", and then answers with "{answer}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint first asks tellme to create the file "([^"]*)" without a reason, then with the content "([^"]*)" and the reason "([^"]*)", and then answers with "([^"]*)"$`, givenProviderCreateFirstReasonlessThenReason)
	})
}

// givenProviderCreateFirstReasonlessThenReason scripts a three-step exchange: a
// reason-less write_file call (REFUSED — the file is not written), then the same
// call carrying a reason (executes), then the answer (round 056 / ADR 0025 D3).
func givenProviderCreateFirstReasonlessThenReason(ctx context.Context, provider, path, content, reason, answer string) error {
	sc := scenarioFrom(ctx)
	path = unescapeText(path)
	content = unescapeText(content)
	reason = unescapeText(reason)
	noReason, _ := json.Marshal(map[string]any{"filepath": path, "content": content})
	withReason, _ := json.Marshal(map[string]any{"filepath": path, "content": content, "reason": reason})
	f := sc.newFake()
	f.Script(
		fakeprovider.Reply{ToolName: "write_file", Arguments: string(noReason)},
		fakeprovider.Reply{ToolName: "write_file", Arguments: string(withReason)},
		fakeprovider.Reply{Answer: unescapeText(answer)},
	)
	sc.scriptedTool = "write_file"
	sc.scriptedAnswer = unescapeText(answer)
	sc.scriptedAnswerSet = true
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

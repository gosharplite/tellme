package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// R028 — Given: the environment prompt log already holds "{prompt}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the environment prompt log already holds "([^"]*)"$`, givenEnvPromptLogHolds)
	})
}

// givenEnvPromptLogHolds arranges the round-028 seed source: one {timestamp,prompt}
// record in the environment-scoped log at
// <TELL_ME_HOME>/output/global_prompts.jsonl (怎麼做 / 權威狀態落地 / 回寫).
func givenEnvPromptLogHolds(ctx context.Context, prompt string) error {
	return appendEnvPromptLog(scenarioFrom(ctx), prompt)
}

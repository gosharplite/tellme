package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// R028 — Given: the shared prompt log has not been created yet
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the shared prompt log has not been created yet$`, givenSharedLogAbsent)
	})
}

// givenSharedLogAbsent ensures the user-global shared log
// (~/.tellme/global_prompts.jsonl) does not exist, so the next `-i` run's
// first-use seed engages (怎麼做 / 權威狀態落地).
func givenSharedLogAbsent(ctx context.Context) error {
	return ensureSharedPromptLogAbsent(scenarioFrom(ctx))
}

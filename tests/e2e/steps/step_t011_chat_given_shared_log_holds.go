package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T011 — Given: the shared prompt log already holds "{prompt}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the shared prompt log already holds "([^"]*)"$`, givenSharedLogHolds)
	})
}

// givenSharedLogHolds appends one {timestamp,prompt} record to the shared log at
// the output/ ROOT of the runtime home (怎麼做 / 權威狀態落地 / 回寫).
func givenSharedLogHolds(ctx context.Context, prompt string) error {
	return appendPromptLog(scenarioFrom(ctx), prompt)
}

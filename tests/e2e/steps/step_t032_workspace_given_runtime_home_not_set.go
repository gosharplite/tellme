package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T032 — Given: the runtime home is not set
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the runtime home is not set$`, givenRuntimeHomeNotSet)
	})
}

// givenRuntimeHomeNotSet ensures TELL_ME_HOME is unset for the run (怎麼做 /
// 權威狀態落地: no runtime home is resolvable).
func givenRuntimeHomeNotSet(ctx context.Context) error {
	scenarioFrom(ctx).homeSet = false
	return nil
}

package steps

import (
	"context"
	"os"

	"github.com/cucumber/godog"
)

// T007 [BDD-RED] — Given: the runtime home holds no skills
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the runtime home holds no skills$`, givenRuntimeHomeHoldsNoSkills)
	})
}

// givenRuntimeHomeHoldsNoSkills (怎麼做 / 權威狀態落地 / 回寫): ensure
// $TELL_ME_HOME/docs/skills does not exist under the runtime home, so the loader
// observes an empty catalog. A fresh scenario home has none, so this is a no-op
// in practice; RemoveAll keeps the arrange idempotent.
func givenRuntimeHomeHoldsNoSkills(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	return os.RemoveAll(sc.homePath("docs/skills"))
}

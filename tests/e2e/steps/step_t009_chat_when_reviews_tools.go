package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T009 [BDD-RED] — When: the operator reviews how the tools have been used
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.When(`^the operator reviews how the tools have been used$`, whenReviewsTools)
	})
}

// whenReviewsTools (怎麼做 / 權威狀態落地 / 回寫): run `tellme --tool-usage`. The
// report is `--version`-class — it needs no -c, no TELL_ME_HOME, no workspace,
// and reads no stdin.
func whenReviewsTools(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	sc.args = []string{"--tool-usage"}
	sc.run()
	return nil
}

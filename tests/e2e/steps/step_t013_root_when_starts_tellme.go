package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T013 — When: the operator starts tellme
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.When(`^the operator starts tellme$`, whenStartsTellme)
	})
}

// whenStartsTellme runs the built tellme binary with no -c, under the current
// environment (怎麼做 / 權威狀態落地 / 回寫: run to completion; capture exit
// code, stdout, stderr).
func whenStartsTellme(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	sc.args = nil
	sc.run()
	return nil
}

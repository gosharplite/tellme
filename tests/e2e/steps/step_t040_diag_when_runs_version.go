package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T040 — When: the operator runs tellme with "--version"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.When(`^the operator runs tellme with "([^"]*)"$`, whenRunsWithFlag)
	})
}

// whenRunsWithFlag runs `tellme {flag}` (怎麼做 / 權威狀態落地 / 回寫: run to
// completion; capture exit code, stdout, stderr).
func whenRunsWithFlag(ctx context.Context, flag string) error {
	sc := scenarioFrom(ctx)
	sc.args = []string{flag}
	sc.run()
	return nil
}

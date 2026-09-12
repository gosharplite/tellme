package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T011 — When: the operator starts a fresh session with "--new"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.When(`^the operator starts a fresh session with "--new"$`, whenStartFreshSession)
	})
}

// whenStartFreshSession runs `tellme --new` (no -c, no positional prompt) under
// the current environment (怎麼做 / 回寫: capture exit code, stdout, stderr, and
// the resulting workspace files).
func whenStartFreshSession(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	sc.args = []string{"--new"}
	sc.run()
	return nil
}

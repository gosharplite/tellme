package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T041 — When: the operator runs tellme's diagnostic
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.When(`^the operator runs tellme's diagnostic$`, whenRunsDiagnostic)
	})
}

// whenRunsDiagnostic runs `tellme -d` (怎麼做 / 權威狀態落地 / 回寫: the report
// is produced regardless of the resolution outcome; capture exit code, stdout,
// stderr).
func whenRunsDiagnostic(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	sc.args = []string{"-d"}
	sc.run()
	return nil
}

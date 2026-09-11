package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T042 — When: the operator runs tellme's diagnostic with "--json"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.When(`^the operator runs tellme's diagnostic with "([^"]*)"$`, whenRunsDiagnosticWithFlag)
	})
}

// whenRunsDiagnosticWithFlag runs `tellme -d {flag}` (怎麼做 / 權威狀態落地 /
// 回寫: the structured report is produced regardless of outcome; capture exit
// code, stdout, stderr).
func whenRunsDiagnosticWithFlag(ctx context.Context, flag string) error {
	sc := scenarioFrom(ctx)
	sc.args = []string{"-d", flag}
	sc.run()
	return nil
}

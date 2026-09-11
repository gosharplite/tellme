package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T007 — When: the operator runs tellme's diagnostic with "--json"
// (usage module: --json is no longer a flag, so `-d --json` is a usage error).
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.When(`^the operator runs tellme's diagnostic with "--json"$`, whenRunsDiagnosticWithRemovedJSON)
	})
}

// whenRunsDiagnosticWithRemovedJSON runs `tellme -d --json`
// (怎麼做 / 權威狀態落地 / 回寫: parsing runs before the diagnostic dispatch; --json
// is not a flag, so the run is a usage error even though -d is present; capture
// exit code, stdout, stderr).
func whenRunsDiagnosticWithRemovedJSON(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	sc.args = []string{"-d", "--json"}
	sc.run()
	return nil
}

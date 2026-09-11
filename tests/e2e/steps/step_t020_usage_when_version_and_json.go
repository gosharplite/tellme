package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T020 (N2) — When: the operator starts tellme with "--version" and the
// unrecognized flag "--json" (usage module).
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.When(`^the operator starts tellme with "--version" and the unrecognized flag "--json"$`, whenVersionAndRemovedJSON)
	})
}

// whenVersionAndRemovedJSON runs `tellme --version --json`
// (怎麼做 / 權威狀態落地 / 回寫: parsing runs before the version path; --json is not a
// flag, so the run is a usage error even though --version is present; capture
// exit code, stdout, stderr).
func whenVersionAndRemovedJSON(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	sc.args = []string{"--version", "--json"}
	sc.run()
	return nil
}

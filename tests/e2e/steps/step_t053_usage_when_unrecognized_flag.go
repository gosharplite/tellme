package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T053 — When: the operator starts tellme pointing at the configuration "{config_path}" with the unrecognized flag "{flag}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.When(`^the operator starts tellme pointing at the configuration "([^"]*)" with the unrecognized flag "([^"]*)"$`, whenStartsWithConfigAndFlag)
	})
}

// whenStartsWithConfigAndFlag runs `tellme -c $TELL_ME_HOME/{config_path} {flag}`
// (怎麼做 / 權威狀態落地 / 回寫: parsing runs before resolution; the unrecognized
// flag is a usage error even with a valid configuration present).
func whenStartsWithConfigAndFlag(ctx context.Context, configPath, flag string) error {
	sc := scenarioFrom(ctx)
	sc.args = []string{"-c", sc.homePath(configPath), flag}
	sc.run()
	return nil
}

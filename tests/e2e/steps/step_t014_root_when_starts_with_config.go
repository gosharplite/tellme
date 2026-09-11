package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T014 — When: the operator starts tellme pointing at the configuration "{config_path}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.When(`^the operator starts tellme pointing at the configuration "([^"]*)"$`, whenStartsWithConfig)
	})
}

// whenStartsWithConfig runs `tellme -c $TELL_ME_HOME/{config_path}` (怎麼做 /
// 權威狀態落地 / 回寫: run to completion; capture exit code, stdout, stderr).
func whenStartsWithConfig(ctx context.Context, configPath string) error {
	sc := scenarioFrom(ctx)
	sc.args = []string{"-c", sc.homePath(configPath)}
	sc.run()
	return nil
}

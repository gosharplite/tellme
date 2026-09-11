package steps

import (
	"context"
	"fmt"
	"os"

	"github.com/cucumber/godog"
)

// T011 — Given: the runtime home is "{home}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the runtime home is "([^"]*)"$`, givenRuntimeHome)
	})
}

// givenRuntimeHome arranges a fresh empty runtime home exposed as TELL_ME_HOME.
// The operator-facing name passed in stands for the arranged home (怎麼做 /
// 權威狀態落地: TELL_ME_HOME is set in the process environment).
func givenRuntimeHome(ctx context.Context, _ string) error {
	sc := scenarioFrom(ctx)
	if sc == nil {
		return fmt.Errorf("no scenario context")
	}
	if err := os.MkdirAll(sc.home, 0o755); err != nil {
		return fmt.Errorf("create runtime home: %w", err)
	}
	sc.homeSet = true
	return nil
}

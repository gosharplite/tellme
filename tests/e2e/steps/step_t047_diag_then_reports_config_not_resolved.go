package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T047 — Then: tellme reports the configuration did not resolve
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme reports the configuration did not resolve$`, thenReportsConfigUnresolved)
	})
}

// thenReportsConfigUnresolved (必查 呈現結果): stdout reports the configuration
// did not resolve; 權威狀態: the reported status matches the arranged setup.
func thenReportsConfigUnresolved(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if !strings.Contains(sc.stdout, "configuration: unresolved") {
		return fmt.Errorf("stdout %q does not report the configuration did not resolve", sc.stdout)
	}
	return nil
}

package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T044 — Then: tellme reports the configuration resolved
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme reports the configuration resolved$`, thenReportsConfigResolved)
	})
}

// thenReportsConfigResolved (必查 呈現結果): stdout reports the configuration
// resolved; 權威狀態: the resolution matches a ready configuration.
func thenReportsConfigResolved(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if !strings.Contains(sc.stdout, "configuration: resolved") {
		return fmt.Errorf("stdout %q does not report the configuration resolved", sc.stdout)
	}
	return nil
}

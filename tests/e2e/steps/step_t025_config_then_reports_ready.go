package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T025 — Then: tellme reports the configuration is ready
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme reports the configuration is ready$`, thenReportsConfigReady)
	})
}

// thenReportsConfigReady (必查 呈現結果): stdout reports the configuration is
// ready; the run validated the file and the effective selected provider.
func thenReportsConfigReady(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if !strings.Contains(sc.stdout, "configuration: ready") {
		return fmt.Errorf("stdout %q does not report the configuration is ready", sc.stdout)
	}
	return nil
}

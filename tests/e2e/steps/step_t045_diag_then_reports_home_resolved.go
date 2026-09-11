package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T045 — Then: tellme reports the runtime home resolved to "{home}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme reports the runtime home resolved to "([^"]*)"$`, thenReportsHomeResolved)
	})
}

// thenReportsHomeResolved (必查 呈現結果): stdout reports the runtime home
// resolved to the arranged TELL_ME_HOME.
func thenReportsHomeResolved(ctx context.Context, _ string) error {
	sc := scenarioFrom(ctx)
	if !strings.Contains(sc.stdout, sc.home) {
		return fmt.Errorf("stdout %q does not report the runtime home %q", sc.stdout, sc.home)
	}
	return nil
}

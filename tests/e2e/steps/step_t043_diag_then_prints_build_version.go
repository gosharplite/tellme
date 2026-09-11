package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/harness"
)

// T043 — Then: tellme prints the build version
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme prints the build version$`, thenPrintsBuildVersion)
	})
}

// thenPrintsBuildVersion (必查 呈現結果): stdout contains the build version
// string. The E2E binary is built with the sentinel, so a missed injection
// (which would leave `dev`) fails here.
func thenPrintsBuildVersion(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if !strings.Contains(sc.stdout, harness.SentinelVersion) {
		return fmt.Errorf("stdout %q does not contain build version %q (a missed -X injection prints the default)", sc.stdout, harness.SentinelVersion)
	}
	return nil
}

package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T030 [BDD-RED] — Then: tellme reports that "{path}" is a directory
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme reports that "([^"]*)" is a directory$`, thenReportsDirectory)
	})
}

// thenReportsDirectory (必查 權威狀態): the read_files result for {path} carries the
// directory error line.
func thenReportsDirectory(ctx context.Context, path string) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	if !strings.Contains(lastToolResult(f), "ERROR: path is a directory, use list_files instead") {
		return fmt.Errorf("the read of %q was not reported as a directory; result=%q", path, lastToolResult(f))
	}
	return nil
}

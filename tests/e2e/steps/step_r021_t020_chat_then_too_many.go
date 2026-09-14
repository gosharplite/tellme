package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T020 [BDD-RED] — Then: tellme reports that too many files were requested
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme reports that too many files were requested$`, thenTooMany)
	})
}

// thenTooMany (必查 權威狀態): the read_files result carries the too-many-files
// message (a non-fatal result, not a tool failure).
func thenTooMany(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	if !strings.Contains(lastToolResult(f), "requested too many files") {
		return fmt.Errorf("the read result did not report too many files; result=%q", lastToolResult(f))
	}
	return nil
}

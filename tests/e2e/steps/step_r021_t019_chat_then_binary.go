package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T019 [BDD-RED] — Then: tellme reports that "{name}" is a binary file that cannot be shown as text
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme reports that "([^"]*)" is a binary file that cannot be shown as text$`, thenBinary)
	})
}

// thenBinary (必查 權威狀態): the read_files result for {name} carries the binary
// marker (the reference wording).
func thenBinary(ctx context.Context, name string) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	if !strings.Contains(lastToolResult(f), "(Binary file, cannot display as text)") {
		return fmt.Errorf("the read of %q was not reported as binary; result=%q", name, lastToolResult(f))
	}
	return nil
}

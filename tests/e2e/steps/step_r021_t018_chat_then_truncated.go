package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T018 [BDD-RED] — Then: the part of "{name}" that tellme read ends with a truncation marker
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the part of "([^"]*)" that tellme read ends with a truncation marker$`, thenTruncated)
	})
}

// thenTruncated (必查 權威狀態): the read_files result for {name} ends with the
// per-file truncation marker. A SUFFIX match (not Contains) distinguishes it from
// the aggregate marker "... (truncated at the read budget)".
func thenTruncated(ctx context.Context, name string) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	result := strings.TrimRight(lastToolResult(f), "\n")
	if !strings.HasSuffix(result, "(truncated)") {
		at := len(result) - 48
		if at < 0 {
			at = 0
		}
		return fmt.Errorf("the read of %q does not end with a truncation marker; tail=%q", name, result[at:])
	}
	return nil
}

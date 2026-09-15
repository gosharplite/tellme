package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T031 [BDD-RED] — Then: the read result frames "{name}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the read result frames "([^"]*)"$`, thenReadFramed)
	})
}

// thenReadFramed (必查 權威狀態): the read_files result carries the per-file header
// line "--- File: {name} ---".
func thenReadFramed(ctx context.Context, name string) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	if !strings.Contains(lastToolResult(f), "--- File: "+name+" ---") {
		return fmt.Errorf("the read result does not frame %q; result=%q", name, lastToolResult(f))
	}
	return nil
}

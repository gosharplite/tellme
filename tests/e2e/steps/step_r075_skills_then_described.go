package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// Round-075 Then: the listing describes a skill with its resolved text. Kept in
// its own file (Zero Shared Edits).
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the listing describes the skill "([^"]*)" as "([^"]*)"$`, thenListingDescribesSkill)
	})
}

// thenListingDescribesSkill (必查 權威狀態): the `list_skills` result carries the
// entry `- {name}: {description} (` — the description resolved to its real text
// (for a block scalar, the folded value), never the bare indicator.
func thenListingDescribesSkill(ctx context.Context, name, description string) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	result := lastToolResult(f)
	want := "- " + name + ": " + description + " ("
	if !strings.Contains(result, want) {
		return fmt.Errorf("the listing does not describe the skill %q as %q; result=%q", name, description, result)
	}
	// The DSL row's 不該發生 clause (fold N-1): a block-scalar description must not
	// be listed as the bare indicator.
	if strings.Contains(result, "- "+name+": >") {
		return fmt.Errorf("the listing shows the bare block-scalar indicator for the skill %q; result=%q", name, result)
	}
	return nil
}

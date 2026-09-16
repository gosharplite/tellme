package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T012 [BDD-RED] — Then: the listing does not include "{text}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the listing does not include "([^"]*)"$`, thenListingExcludesText)
	})
}

// thenListingExcludesText (必查 權威狀態): the `list_skills` result contains NO
// occurrence of {text} — a non-skill file (a skill's rules/*.md, STANDARDS.md)
// must not appear as a skill.
func thenListingExcludesText(ctx context.Context, text string) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	result := lastToolResult(f)
	if strings.Contains(result, text) {
		return fmt.Errorf("the listing includes %q; result=%q", text, result)
	}
	return nil
}

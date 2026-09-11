package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T048 — Then: tellme reports the reason the configuration did not resolve
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme reports the reason the configuration did not resolve$`, thenReportsUnresolvedReason)
	})
}

// thenReportsUnresolvedReason (必查 呈現結果): stdout names the unresolved
// category (one of the pinned reasons) consistent with the arranged setup.
func thenReportsUnresolvedReason(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	for _, category := range unresolvedCategories {
		if strings.Contains(sc.stdout, category) {
			return nil
		}
	}
	return fmt.Errorf("stdout %q names no unresolved category (want one of %v)", sc.stdout, unresolvedCategories)
}

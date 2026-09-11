package steps

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cucumber/godog"
)

// T050 — Then: tellme emits the unresolved status as structured output
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme emits the unresolved status as structured output$`, thenEmitsUnresolvedJSON)
	})
}

// thenEmitsUnresolvedJSON (必查 呈現結果): stdout parses as the pinned JSON
// object with status unresolved and reason equal to a pinned category.
func thenEmitsUnresolvedJSON(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	var obj diagnosticJSON
	if err := json.Unmarshal([]byte(sc.stdout), &obj); err != nil {
		return fmt.Errorf("stdout is not the pinned JSON object: %v; got %q", err, sc.stdout)
	}
	if obj.Status != "unresolved" {
		return fmt.Errorf("status = %q, want %q", obj.Status, "unresolved")
	}
	for _, category := range unresolvedCategories {
		if obj.Reason == category {
			return nil
		}
	}
	return fmt.Errorf("reason = %q, want one of %v", obj.Reason, unresolvedCategories)
}

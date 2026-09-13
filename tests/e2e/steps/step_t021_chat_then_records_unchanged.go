package steps

import "github.com/cucumber/godog"

// T021 — Then: the earlier conversation records are unchanged
//
// Skeleton (round-008 T007): registered + implemented in Phase 3 (BDD-RED).
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		_ = ctx // TODO(T021): ctx.Then(...) built from the DSL row's StepDef 實作語意
	})
}

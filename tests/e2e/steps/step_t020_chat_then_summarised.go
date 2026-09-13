package steps

import "github.com/cucumber/godog"

// T020 — Then: tellme summarised the earlier conversation using its summarise tool
//
// Skeleton (round-008 T007): registered + implemented in Phase 3 (BDD-RED).
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		_ = ctx // TODO(T020): ctx.Then(...) built from the DSL row's StepDef 實作語意
	})
}

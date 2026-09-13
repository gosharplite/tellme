package steps

import "github.com/cucumber/godog"

// T014 — Then: tellme read "{path}" using its read_files tool
//
// Skeleton (round-008 T007): registered + implemented in Phase 3 (BDD-RED).
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		_ = ctx // TODO(T014): ctx.Then(...) built from the DSL row's StepDef 實作語意
	})
}

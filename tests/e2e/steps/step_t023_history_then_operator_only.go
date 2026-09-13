package steps

import "github.com/cucumber/godog"

// T023 — Then: tellme lists only the operator's messages
//
// Skeleton (round-008 T007): registered + implemented in Phase 3 (BDD-RED).
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		_ = ctx // TODO(T023): ctx.Then(...) built from the DSL row's StepDef 實作語意
	})
}

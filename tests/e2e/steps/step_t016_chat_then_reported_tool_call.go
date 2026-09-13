package steps

import "github.com/cucumber/godog"

// T016 — Then: the run reported the tool call "{tool}" on its diagnostic output
//
// Skeleton (round-008 T007): registered + implemented in Phase 3 (BDD-RED).
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		_ = ctx // TODO(T016): ctx.Then(...) built from the DSL row's StepDef 實作語意
	})
}

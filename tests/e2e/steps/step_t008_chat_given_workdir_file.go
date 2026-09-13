package steps

import "github.com/cucumber/godog"

// T008 — Given: the working directory contains a file "{name}" whose text is "{content}"
//
// Skeleton (round-008 T007): registered + implemented in Phase 3 (BDD-RED).
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		_ = ctx // TODO(T008): ctx.Given(...) built from the DSL row's StepDef 實作語意
	})
}

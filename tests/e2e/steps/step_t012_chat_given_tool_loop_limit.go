package steps

import "github.com/cucumber/godog"

// T012 — Given: the tool-loop limit is "{limit}"
//
// Skeleton (round-008 T007): registered + implemented in Phase 3 (BDD-RED).
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		_ = ctx // TODO(T012): ctx.Given(...) built from the DSL row's StepDef 實作語意
	})
}

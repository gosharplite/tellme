package steps

import "github.com/cucumber/godog"

// T011 — Given: a configured provider "{provider}" whose endpoint asks for a tool that is not available
//
// Skeleton (round-008 T007): registered + implemented in Phase 3 (BDD-RED).
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		_ = ctx // TODO(T011): ctx.Given(...) built from the DSL row's StepDef 實作語意
	})
}

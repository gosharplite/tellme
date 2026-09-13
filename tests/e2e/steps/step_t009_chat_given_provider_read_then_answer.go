package steps

import "github.com/cucumber/godog"

// T009 — Given: a configured provider "{provider}" whose endpoint asks tellme to read "{path}" and then answers with "{answer}"
//
// Skeleton (round-008 T007): registered + implemented in Phase 3 (BDD-RED).
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		_ = ctx // TODO(T009): ctx.Given(...) built from the DSL row's StepDef 實作語意
	})
}

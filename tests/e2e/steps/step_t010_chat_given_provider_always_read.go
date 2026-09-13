package steps

import "github.com/cucumber/godog"

// T010 — Given: a configured provider "{provider}" whose endpoint always asks tellme to read "{path}"
//
// Skeleton (round-008 T007): registered + implemented in Phase 3 (BDD-RED).
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		_ = ctx // TODO(T010): ctx.Given(...) built from the DSL row's StepDef 實作語意
	})
}

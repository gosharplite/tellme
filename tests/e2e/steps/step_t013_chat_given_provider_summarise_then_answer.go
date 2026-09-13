package steps

import "github.com/cucumber/godog"

// T013 — Given: a configured provider "{provider}" whose endpoint asks tellme to summarise the conversation and then answers with "{answer}"
//
// Skeleton (round-008 T007): registered + implemented in Phase 3 (BDD-RED).
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		_ = ctx // TODO(T013): ctx.Given(...) built from the DSL row's StepDef 實作語意
	})
}

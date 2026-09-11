// Package steps holds the godog step definitions for the tellme CLI E2E suite.
//
// The registration mechanism is deliberately split to enable zero-shared-edit
// parallel authoring: every DSL step file lives in its own file and appends its
// own registration function to `registrars` from its init(). RegisterAll drains
// that slice, so adding a step never edits a shared file.
package steps

import "github.com/cucumber/godog"

// registrars collects per-file registration functions. Each step file appends
// its own register function in an init(); no file edits this slice directly.
var registrars []func(*godog.ScenarioContext)

// RegisterAll wires the scenario lifecycle hooks and every registered step
// definition onto the scenario context. Called once per scenario by the suite.
func RegisterAll(ctx *godog.ScenarioContext) {
	ctx.Before(beforeScenario)
	ctx.After(afterScenario)
	for _, register := range registrars {
		register(ctx)
	}
}

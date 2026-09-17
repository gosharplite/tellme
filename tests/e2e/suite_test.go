// Package e2e is the godog test suite that drives the built `tellme` binary
// end-to-end over the executable CLI contract in specs/truth/features/cli.
package e2e

import (
	"testing"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/steps"
)

// TestFeatures is the godog suite entry. It loads the interface Gherkin from
// the truth tree and wires every scenario through the steps package, which owns
// scenario state and per-task self-registration.
//
// Round 040 review TD-1 adds `Strict: true` here IN THE IMPLEMENTATION HALF,
// together with the four new stepdefs. godog's Strict default is false, so an
// undefined step is silently reported while the suite exits 0 — but `make test`
// is `go test ./...` (which includes this package), so enabling Strict on a
// plan-half branch would make `dev` red on every run until the stepdefs land.
// The implement half therefore lands the flag with the steps, witnessed by the
// FAIL-then-PASS transition. See the round-040 plan's task directives.
func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		Name:                "cli",
		ScenarioInitializer: func(ctx *godog.ScenarioContext) { steps.RegisterAll(ctx) },
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../specs/truth/features/cli"},
			TestingT: t,
			// Round 040 TD-1: undefined steps FAIL the suite (they were previously
			// reported-and-ignored). Landed here, with the round's stepdefs, so the
			// FAIL→PASS transition witnesses the new Examples.
			Strict: true,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("godog suite reported a non-zero status")
	}
}

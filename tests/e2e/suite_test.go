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
// Strict: true (round 040 review TD-1) — godog's Strict default is false, so an
// undefined/pending/ambiguous step was silently reported in the pretty output
// while the suite exited 0. That let a round's new truth Examples merge without
// ever executing (round 040 shipped 7 new steps). Strict makes those fail
// loudly: a plan-half branch carrying not-yet-implemented steps is red until
// /axb-implement defines them — the honest state, and it makes the round's own
// "the E2E suite is green" success criterion mechanically checkable.
func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		Name:                "cli",
		ScenarioInitializer: func(ctx *godog.ScenarioContext) { steps.RegisterAll(ctx) },
		Options: &godog.Options{
			Format:   "pretty",
			Strict:   true,
			Paths:    []string{"../../specs/truth/features/cli"},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("godog suite reported a non-zero status")
	}
}

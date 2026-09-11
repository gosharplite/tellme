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
func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		Name:                "cli",
		ScenarioInitializer: func(ctx *godog.ScenarioContext) { steps.RegisterAll(ctx) },
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../specs/truth/features/cli"},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("godog suite reported a non-zero status")
	}
}

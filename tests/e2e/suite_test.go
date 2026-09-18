// Package e2e is the godog test suite that drives the built `tellme` binary
// end-to-end over the executable CLI contract in specs/truth/features/cli.
package e2e

import (
	"flag"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/steps"
)

// e2eDefaultConcurrency is the number of scenarios the suite runs in parallel
// by default (round 055, ADR 0024 D1). Four was chosen from measurement: the
// suite's isolation (a per-scenario temp TELL_ME_HOME *and* HOME, immutable
// package-level step state, a single once-built binary) makes it safe, and the
// measured wall-clock went ~50.8s serial -> ~16.2s at 4 (3.0x); 8 gained only a
// further ~1.3s while doubling the concurrent child count. Override with
// TELL_ME_E2E_CONCURRENCY (a lower value is the escape hatch on a
// resource-constrained CI — ADR 0024 RF-055-3).
const e2eDefaultConcurrency = 4

// e2eFeaturesRoot is the whole executable contract. A gate run always covers it
// (ADR 0024 D2/D4); only `make test-fast` narrows it, and never to the root.
const e2eFeaturesRoot = "../../specs/truth/features/cli"

// e2ePathsOverride is the `-godog.paths` selection (comma-separated) used by
// `make test-fast` to run a SUBSET of the contract for a fast inner loop
// (round 055, ADR 0024 D3). It is registered on the stdlib flag set in init()
// so `go test -args -godog.paths=<subset>` reaches the suite: godog's own
// TestSuite/Options path never binds its flags, so the harness owns the flag
// under the same name and semantics (a comma-separated feature path list —
// godog v0.16.0 has no name filter, and tags would edit specs/truth/**, a
// truth-owner change; ADR 0024 RF-055-1). Empty = the full contract (the gate).
var e2ePathsOverride string

func init() {
	flag.StringVar(&e2ePathsOverride, "godog.paths", "",
		"comma-separated feature paths to run (SUBSET — not the gate); empty runs the whole contract")
}

// e2eConcurrency resolves the scenario concurrency: the TELL_ME_E2E_CONCURRENCY
// value when it parses to >= 1, else e2eDefaultConcurrency (ADR 0024 D1).
func e2eConcurrency() int {
	if v := os.Getenv("TELL_ME_E2E_CONCURRENCY"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 1 {
			return n
		}
	}
	return e2eDefaultConcurrency
}

// e2ePaths resolves the feature paths to execute: the -godog.paths selection
// when given (a SUBSET), else the whole contract root (the gate — ADR 0024 D4).
func e2ePaths() []string {
	var paths []string
	for _, p := range strings.Split(e2ePathsOverride, ",") {
		if p = strings.TrimSpace(p); p != "" {
			paths = append(paths, p)
		}
	}
	if len(paths) == 0 {
		return []string{e2eFeaturesRoot}
	}
	return paths
}

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
//
// Round 055 (ADR 0024) adds `Concurrency` (on by default, env-overridable) and
// the `-godog.paths` subset selection. The gate's SCOPE is unchanged: without a
// selection, every Example under the root is still executed, and Strict /
// -count=1 are untouched. `make test-fast` is a developer convenience and is
// never substituted for the gate.
func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		Name:                "cli",
		ScenarioInitializer: func(ctx *godog.ScenarioContext) { steps.RegisterAll(ctx) },
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    e2ePaths(),
			TestingT: t,
			// Round 040 TD-1: undefined steps FAIL the suite (they were previously
			// reported-and-ignored). Landed here, with the round's stepdefs, so the
			// FAIL→PASS transition witnesses the new Examples.
			Strict: true,
			// Round 055 (ADR 0024 D1): scenarios run in parallel by default.
			Concurrency: e2eConcurrency(),
		},
	}
	if suite.Run() != 0 {
		t.Fatal("godog suite reported a non-zero status")
	}
}

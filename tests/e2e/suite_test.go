// Package e2e is the godog test suite that drives the built `tellme` binary
// end-to-end over the executable CLI contract in specs/truth/features/cli.
package e2e

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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
// measured wall-clock went ~50.7s serial -> ~16.9s at 4 (~3.0x); 8 gained only
// a further ~1.3s while doubling the concurrent child count. The gain is
// wait-overlap (the serial floor is dominated by never-answering MCP scenarios
// plus the sleep-based timeout legs), not CPU parallelism — which is why 4
// survives a small CI. Override with TELL_ME_E2E_CONCURRENCY (a lower value is
// the escape hatch on a resource-constrained CI — ADR 0024 RF-055-3).
const e2eDefaultConcurrency = 4

// e2eFeaturesRoot is the whole executable contract. A gate run always covers it
// (ADR 0024 D2/D4); only a provided -godog.paths selection narrows it, and the
// guard below refuses a selection that resolves back to it.
const e2eFeaturesRoot = "../../specs/truth/features/cli"

// e2ePathsOverride is the `-godog.paths` selection (comma-separated) used by
// `make test-fast` to run a SUBSET of the contract for a fast inner loop
// (round 055, ADR 0024 D3). It is registered on the stdlib flag set in init()
// so `go test -args -godog.paths=<subset>` reaches the suite: godog's own
// TestSuite/Options path never binds its flags, so the harness owns the flag
// under the same name and semantics (a comma-separated feature path list —
// godog v0.16.0 has no name filter, and tags would edit specs/truth/**, a
// truth-owner change; ADR 0024 RF-055-1). Empty = the full contract (the gate).
//
// Caveats (ADR 0024 D3, review R-055-x): only `paths` is registered (the other
// `godog.*` flags are undefined, so passing one is a hard usage error — the
// concurrency seam is TELL_ME_E2E_CONCURRENCY, NOT `-godog.concurrency`); and
// `-args` is application-wide, so a repo-wide `go test ./... -args -godog.paths`
// would fail every other package (a subset is a tests/e2e-only invocation).
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
	paths := splitFeaturePaths(e2ePathsOverride)
	if len(paths) == 0 {
		return []string{e2eFeaturesRoot}
	}
	return paths
}

// splitFeaturePaths splits a comma-separated selection, trimming blanks.
func splitFeaturePaths(selection string) []string {
	var paths []string
	for _, p := range strings.Split(selection, ",") {
		if p = strings.TrimSpace(p); p != "" {
			paths = append(paths, p)
		}
	}
	return paths
}

// guardSelection enforces the round-055 invariant (ADR 0024 D3, FR-005, plan.md
// I-3): a PROVIDED selection that resolves to the whole contract is refused —
// "a subset selects for convenience and NEVER excludes from the gate". The gate
// itself (no selection) is legitimate and is never refused, and a genuine
// subset returns nil. The decision is made on RESOLVED paths (the set of
// `.feature` files each selection contains), never on synthesized strings, so
// `../cli` traversal and an all-modules list are both caught (review B-055-1).
func guardSelection(selection string) error {
	paths := splitFeaturePaths(selection)
	if len(paths) == 0 {
		return nil // the gate: no selection supplied
	}
	selected, err := featureSet(paths)
	if err != nil {
		return err
	}
	if len(selected) == 0 {
		return errors.New("the selection contains no .feature files")
	}
	root, err := featureSet([]string{e2eFeaturesRoot})
	if err != nil {
		return err
	}
	if len(selected) == len(root) {
		same := true
		for f := range root {
			if _, ok := selected[f]; !ok {
				same = false
				break
			}
		}
		if same {
			return fmt.Errorf("the selection resolves to the WHOLE contract (%s) — that is the gate; run without -godog.paths (or `make test`)", e2eFeaturesRoot)
		}
	}
	return nil
}

// featureSet resolves feature paths to the set of canonical absolute `.feature`
// file paths they contain, so equality/containment is decided on what actually
// runs. A missing path is an error (fail loud).
func featureSet(paths []string) (map[string]struct{}, error) {
	out := make(map[string]struct{})
	for _, p := range paths {
		abs, err := filepath.Abs(p)
		if err != nil {
			return nil, err
		}
		if resolved, err := filepath.EvalSymlinks(abs); err == nil {
			abs = resolved
		}
		info, err := os.Stat(abs)
		if err != nil {
			return nil, fmt.Errorf("feature path %q is not available", p)
		}
		if !info.IsDir() {
			out[abs] = struct{}{}
			continue
		}
		walkErr := filepath.WalkDir(abs, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && strings.HasSuffix(path, ".feature") {
				if resolved, err := filepath.EvalSymlinks(path); err == nil {
					path = resolved
				}
				out[path] = struct{}{}
			}
			return nil
		})
		if walkErr != nil {
			return nil, walkErr
		}
	}
	return out, nil
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
// -count=1 are untouched. A provided selection is guarded here so that BOTH
// entry points (`make test-fast` and a hand-typed `go test -args`) obey the
// never-the-gate invariant (review B-055-1).
func TestFeatures(t *testing.T) {
	if e2ePathsOverride != "" {
		if err := guardSelection(e2ePathsOverride); err != nil {
			t.Fatalf("e2e subset refused: %v", err)
		}
		fmt.Fprintf(os.Stderr, "SUBSET — NOT THE GATE: %s (the gate: `make test`, or no -godog.paths)\n", e2ePathsOverride)
	}
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

//go:build arch

// The layer-discipline gate. Everything lives in this build-tagged test file so
// it runs only under `-tags=arch` (the Makefile `verify-architecture` target)
// and never in the default `go test ./...` run. See docs/decisions/0011-layer-
// discipline-gate.md and specs/truth/techstack.md for the rule and its policy.
package arch

import (
	"bytes"
	"flag"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"sort"
	"strings"
	"testing"
)

// updateBaseline regenerates the committed baseline from the gate's own output
// (N-1). Passed through the test binary: `go test ... -args -update-baseline`;
// exposed as the `make verify-architecture-update` target.
var updateBaseline = flag.Bool("update-baseline", false,
	"regenerate tools/arch/baseline.txt from the gate's own output")

// modulePath is the module's import-path prefix; everything else is stdlib or
// a third-party import and is out of scope for the layer rule.
const modulePath = "github.com/gosharplite/tellme/"

// listFormat is the `go list` template. It carries three import sets: production
// imports, in-package test imports (`.TestImports`), and external test-package
// imports (`.XTestImports`). Test imports ARE evaluated (spec.md Q2: "`_test.go`
// imports in `internal/**` are also governed") — `.Imports` alone would miss a
// test-only upward import (review Fold 1).
const listFormat = `{{.ImportPath}}|{{join .Imports " "}}|{{join .TestImports " "}}|{{join .XTestImports " "}}`

// crossTargets mirrors the Makefile's CROSS_TARGETS. The gate evaluates the
// UNION over every supported target so an OS-gated illegal import cannot hide
// (ADR 0011 D5 / issue #93).
var crossTargets = []struct{ goos, goarch string }{
	{"linux", "amd64"},
	{"linux", "arm64"},
	{"darwin", "amd64"},
	{"darwin", "arm64"},
}

// droppedBuildEnv is the build-context env the child `go list` must NOT inherit
// (an ambient export must not be able to redden the gate). Everything else —
// notably PATH/HOME/GOPATH/GOMODCACHE/GOCACHE, which carry the warm module cache
// — is preserved (ADR 0011 D5, review F-1).
//
// Defence-in-depth: the Makefile's hermetic `export`/`unexport` block (ADR 0012)
// is the PRIMARY owner for `make`-launched invocations; this filter covers the
// gate's documented DIRECT invocation (`go test -count=1 -tags=arch … ./tools/arch`),
// which bypasses `make`. The two variable sets MUST NOT drift silently.
var droppedBuildEnv = map[string]bool{
	"GOOS":         true,
	"GOARCH":       true,
	"GOARM":        true,
	"CGO_ENABLED":  true,
	"GOFLAGS":      true,
	"GO111MODULE":  true,
	"GOEXPERIMENT": true,
	"GOWORK":       true,
}

// tier ranks a governed package (low -> high). ok=false means "internal but
// unranked" (a RULE-D violation). Packages outside internal/ are exempt and are
// never passed to this function.
func tier(rel string) (int, bool) {
	switch {
	case rel == "internal/config" || rel == "internal/home":
		return 1, true
	case rel == "internal/agent" || strings.HasPrefix(rel, "internal/agent/"):
		return 4, true
	case rel == "internal/ui" || strings.HasPrefix(rel, "internal/ui/"):
		return 5, true
	case rel == "internal/cli":
		return 6, true
	case strings.HasPrefix(rel, "internal/domain/"):
		return 0, true
	case strings.HasPrefix(rel, "internal/app/"):
		return 2, true
	case strings.HasPrefix(rel, "internal/infrastructure/"):
		return 3, true
	}
	return 0, false
}

func isInternal(rel string) bool { return strings.HasPrefix(rel, "internal/") }

func isApplicationTier(rel string) bool {
	return rel == "internal/cli" || strings.HasPrefix(rel, "internal/app/")
}

// rel strips the module prefix; ok=false for stdlib / third-party imports.
func rel(importPath string) (string, bool) {
	if !strings.HasPrefix(importPath, modulePath) {
		return "", false
	}
	return strings.TrimPrefix(importPath, modulePath), true
}

// violation reports whether src -> dst breaks the two-part predicate (ADR 0011
// D2): RULE-A (upward), RULE-B (application -> infrastructure), RULE-C (domain
// purity). Both tiers are already known-ranked.
func violation(src string, srcTier int, dst string, dstTier int) bool {
	switch {
	case dstTier > srcTier:
		return true
	case isApplicationTier(src) && strings.HasPrefix(dst, "internal/infrastructure/"):
		return true
	case strings.HasPrefix(src, "internal/domain/") && !strings.HasPrefix(dst, "internal/domain/"):
		return true
	}
	return false
}

// evaluate returns the sorted, canonical violation lines for the graph.
func evaluate(graph map[string]map[string]bool) []string {
	set := map[string]bool{}
	for _, src := range slices.Sorted(maps.Keys(graph)) {
		if !isInternal(src) {
			continue
		}
		st, srcRanked := tier(src)
		if !srcRanked {
			set[src+" -> (unranked governed package)"] = true
			continue
		}
		for _, dst := range slices.Sorted(maps.Keys(graph[src])) {
			if !isInternal(dst) {
				continue
			}
			dt, dstRanked := tier(dst)
			if !dstRanked {
				set[dst+" -> (unranked governed package)"] = true
				continue
			}
			if violation(src, st, dst, dt) {
				set[src+" -> "+dst] = true
			}
		}
	}
	out := make([]string, 0, len(set))
	for v := range set {
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}

// governed reports whether a package is ranked (a governed graph node).
func governed(rel string) bool {
	if !isInternal(rel) {
		return false
	}
	_, ok := tier(rel)
	return ok
}

// cycles returns the strongly-connected components larger than one (Tarjan) over
// the governed ranked graph. Cycles have no baseline — the count must be 0.
func cycles(graph map[string]map[string]bool) [][]string {
	nodes := map[string]bool{}
	adj := map[string][]string{}
	for src, dsts := range graph {
		if !governed(src) {
			continue
		}
		nodes[src] = true
		for dst := range dsts {
			if !governed(dst) {
				continue
			}
			nodes[dst] = true
			adj[src] = append(adj[src], dst)
		}
	}

	index := 0
	idx := map[string]int{}
	low := map[string]int{}
	onStack := map[string]bool{}
	var stack []string
	var out [][]string

	var strongconnect func(v string)
	strongconnect = func(v string) {
		idx[v] = index
		low[v] = index
		index++
		stack = append(stack, v)
		onStack[v] = true
		for _, w := range adj[v] {
			if _, seen := idx[w]; !seen {
				strongconnect(w)
				low[v] = min(low[v], low[w])
			} else if onStack[w] {
				low[v] = min(low[v], idx[w])
			}
		}
		if low[v] == idx[v] {
			var comp []string
			for {
				w := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				onStack[w] = false
				comp = append(comp, w)
				if w == v {
					break
				}
			}
			if len(comp) > 1 {
				sort.Strings(comp)
				out = append(out, comp)
			}
		}
	}
	for _, n := range slices.Sorted(maps.Keys(nodes)) {
		if _, seen := idx[n]; !seen {
			strongconnect(n)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i][0] < out[j][0] })
	return out
}

// moduleRoot walks up from the test process's CWD (a Go test's CWD is its
// package directory, inside the module) to the directory holding go.mod. CWD is
// used first so the resolution survives `-trimpath` (which rewrites
// runtime.Caller's path to a module-relative form); the caller path is a
// fallback.
func moduleRoot(t *testing.T) string {
	t.Helper()
	var starts []string
	if wd, err := os.Getwd(); err == nil {
		starts = append(starts, wd)
	}
	if _, file, _, ok := runtime.Caller(0); ok {
		starts = append(starts, filepath.Dir(file))
	}
	for _, start := range starts {
		dir := start
		for {
			if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
				return dir
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	t.Fatalf("go.mod not found above the test CWD or the guard's source path")
	return ""
}

// childEnv filters the inherited environment for one target. Build-context
// inputs are NEUTRALISED with explicit non-empty values (deleting a key does not
// neutralise a persisted `go env -w` setting — Go falls back to the env file for
// unset AND empty values, review F-2): GOFLAGS=-mod=readonly, GO111MODULE=on,
// GOWORK=off. The warm-cache/runtime variables (PATH/HOME/GOPATH/GOMODCACHE/
// GOCACHE) are preserved (ADR 0011 D5, review F-1).
func childEnv(goos, goarch string) []string {
	env := make([]string, 0, len(os.Environ())+9)
	for _, kv := range os.Environ() {
		key, _, _ := strings.Cut(kv, "=")
		if droppedBuildEnv[key] {
			continue
		}
		env = append(env, kv)
	}
	return append(env,
		"GOOS="+goos,
		"GOARCH="+goarch,
		"CGO_ENABLED=0",
		"GOFLAGS=-mod=readonly",
		"GO111MODULE=on",
		"GOWORK=off",
	)
}

func runGoList(t *testing.T, root, goos, goarch string) string {
	t.Helper()
	cmd := exec.Command("go", "list", "-f", listFormat, "./...")
	cmd.Dir = root
	cmd.Env = childEnv(goos, goarch)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("go list (%s/%s) failed: %v\n%s", goos, goarch, err, stderr.String())
	}
	if strings.TrimSpace(stdout.String()) == "" {
		t.Fatalf("go list (%s/%s) produced no packages (empty graph — refusing a vacuous pass)", goos, goarch)
	}
	return stdout.String()
}

// enumerate returns, for the union over every supported target, two edge sets:
// the merged graph (production + in-package test + external test-package
// imports) that the RULE evaluates, and the **production-only** graph that the
// acyclicity assertion evaluates. The rule governs test imports too (spec.md Q2);
// the cycle property is about code that must compile — Go permits cycles that
// exist only in tests, so a legal same-tier mutual test import must not fail the
// SCC (review Fold 1).
func enumerate(t *testing.T, root string) (merged, prod map[string]map[string]bool) {
	t.Helper()
	merged = map[string]map[string]bool{}
	prod = map[string]map[string]bool{}
	for _, tg := range crossTargets {
		out := runGoList(t, root, tg.goos, tg.goarch)
		for _, line := range strings.Split(out, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			pkg, rest, _ := strings.Cut(line, "|")
			rp, ok := rel(pkg)
			if !ok {
				continue
			}
			if merged[rp] == nil {
				merged[rp] = map[string]bool{}
				prod[rp] = map[string]bool{}
			}
			// Field 1 is production imports; fields 2/3 are in-package and
			// external test-package imports. A test file's import is an edge from
			// the package it lives in (Fold 1).
			imps, rest, _ := strings.Cut(rest, "|")
			testImps, xTestImps, _ := strings.Cut(rest, "|")
			for i, group := range []string{imps, testImps, xTestImps} {
				for _, imp := range strings.Fields(group) {
					rd, ok := rel(imp)
					if !ok {
						continue
					}
					merged[rp][rd] = true
					if i == 0 { // production edge
						prod[rp][rd] = true
					}
				}
			}
		}
	}
	return merged, prod
}

// readBaseline parses the committed baseline. An absent, unreadable, or
// unparsable baseline is a FAILURE — never "no baseline configured" (N-2). A
// well-formed baseline with no entries is allowed here; whether it is acceptable
// depends on the violation set (assertNoNewOrStale / the caller) — a genuine
// zero-violation endpoint must be green (review F-1).
func readBaseline(t *testing.T, path string) []string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("baseline unreadable (%s): %v — an absent/unreadable baseline MUST fail, never be treated as 'no baseline configured'", path, err)
	}
	var lines []string
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if !strings.Contains(line, " -> ") {
			t.Fatalf("malformed baseline line %q in %s (want `<src> -> <dst>`)", line, path)
		}
		lines = append(lines, line)
	}
	sort.Strings(lines)
	return lines
}

// writeBaseline writes the generated baseline (used by -update-baseline / N-1).
func writeBaseline(t *testing.T, path string, violations []string) {
	t.Helper()
	var b strings.Builder
	b.WriteString("# tools/arch/baseline.txt — layer-discipline gate baseline (ADR 0011).\n")
	b.WriteString("#\n")
	b.WriteString("# Generated from the gate's own output — NEVER hand-edit. Regenerate with:\n")
	b.WriteString("#   make verify-architecture-update\n")
	b.WriteString("#   (= go test -count=1 -tags=arch -run TestVerifyRealArchitecture ./tools/arch -args -update-baseline)\n")
	b.WriteString("#\n")
	b.WriteString("# Run the gate directly as:\n")
	b.WriteString("#   go vet -tags=arch ./tools/arch\n")
	b.WriteString("#   go test -count=1 -tags=arch -run TestVerifyRealArchitecture ./tools/arch\n")
	b.WriteString("#\n")
	b.WriteString("# A line `<src> -> <dst>` is a known, baselined RULE-A/B/C violation. This is a\n")
	b.WriteString("# ratchet that only shrinks: a violation not listed here FAILS the gate, and a\n")
	b.WriteString("# listed line that no longer violates (stale) FAILS the gate too.\n")
	for _, v := range violations {
		b.WriteString(v)
		b.WriteString("\n")
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		t.Fatalf("cannot write baseline %s: %v", path, err)
	}
}

// assertGraphEnumerated asserts the enumeration itself (B-2): the whole module,
// not just the guard's own subtree.
func assertGraphEnumerated(t *testing.T, graph map[string]map[string]bool) {
	t.Helper()
	const minPackages = 20
	if len(graph) < minPackages {
		t.Fatalf("enumeration produced only %d package(s); expected the whole module (>= %d) — is the enumeration anchored to the module root?", len(graph), minPackages)
	}
	for _, want := range []string{"internal/cli", "internal/agent", "internal/ui", "internal/ui/tui/prompt", "internal/domain/llm"} {
		if _, ok := graph[want]; !ok {
			t.Fatalf("enumeration is missing known governed package %q — the graph is not the whole module", want)
		}
	}
}

// assertNoUnrankedGoverned asserts RULE-D coverage: every internal package must
// match a tier (default-deny never treats "unknown" as "allowed").
func assertNoUnrankedGoverned(t *testing.T, graph map[string]map[string]bool) {
	t.Helper()
	var unranked []string
	for src := range graph {
		if isInternal(src) {
			if _, ok := tier(src); !ok {
				unranked = append(unranked, src)
			}
		}
	}
	if len(unranked) > 0 {
		sort.Strings(unranked)
		t.Fatalf("internal package(s) matching no tier (RULE-D default-deny): %v — add them to the tier table", unranked)
	}
}

// assertNoNewOrStale applies the ratchet: any violation not in the baseline is a
// new violation, and any baseline line that no longer violates is stale.
func assertNoNewOrStale(t *testing.T, path string, violations, baseline []string) {
	t.Helper()
	inBase := make(map[string]bool, len(baseline))
	for _, b := range baseline {
		inBase[b] = true
	}
	got := make(map[string]bool, len(violations))
	for _, v := range violations {
		got[v] = true
	}
	var fresh, stale []string
	for _, v := range violations {
		if !inBase[v] {
			fresh = append(fresh, v)
		}
	}
	for _, b := range baseline {
		if !got[b] {
			stale = append(stale, b)
		}
	}
	if len(fresh) > 0 {
		t.Errorf("layer-discipline gate: %d new violation(s) not in the baseline:\n  %s", len(fresh), strings.Join(fresh, "\n  "))
	}
	if len(stale) > 0 {
		t.Errorf("layer-discipline gate: %d stale baseline entr(ies) in %s — remove them from the baseline:\n  %s", len(stale), path, strings.Join(stale, "\n  "))
	}
}

// selfTestPredicate unit-tests the two-part predicate on a synthetic graph,
// covering RULE-A/B/C/D and the legal boundaries.
func selfTestPredicate(t *testing.T) {
	t.Helper()
	graph := map[string]map[string]bool{
		"internal/agent":            {"internal/ui": true},
		"internal/app/suggestions":  {"internal/infrastructure/tools": true},
		"internal/cli":              {"internal/infrastructure/history": true},
		"internal/domain/history":   {"internal/config": true},
		"internal/domain/llm":       {"internal/domain/history": true},
		"internal/home":             {"internal/domain/metrics": true},
		"internal/infrastructure/x": {"internal/config": true},
		"internal/infrastructure/z": {"internal/weird": true},
		"internal/ui/tui/prompt":    {"internal/domain/llm": true},
		"internal/weird":            {"internal/domain/llm": true},
	}
	want := []string{
		"internal/agent -> internal/ui",
		"internal/app/suggestions -> internal/infrastructure/tools",
		"internal/cli -> internal/infrastructure/history",
		"internal/domain/history -> internal/config",
		"internal/weird -> (unranked governed package)",
	}
	if got := evaluate(graph); !reflect.DeepEqual(got, want) {
		t.Fatalf("predicate self-test failed:\n got  %v\n want %v", got, want)
	}
}

// TestVerifyRealArchitecture is the gate. It runs all three properties —
// enumeration, ranking/baseline diff, acyclicity — and asserts they all ran, so
// `-run TestVerifyRealArchitecture` alone cannot silently skip one (N-3).
func TestVerifyRealArchitecture(t *testing.T) {
	root := moduleRoot(t)

	properties := 0

	// Property 1 — enumeration (B-2).
	graph, prodGraph := enumerate(t, root)
	assertGraphEnumerated(t, graph)
	properties++

	// The predicate + tier-table coverage (default-deny) on synthetic input.
	selfTestPredicate(t)
	assertNoUnrankedGoverned(t, graph)

	// Property 2 — ranking + baseline diff (rule evaluated on the merged graph,
	// so test imports are governed).
	violations := evaluate(graph)
	properties++

	// Property 3 — acyclicity (ADR 0011 D8 / issue #93 AC4), on the
	// **production-only** union: Go permits test-only cycles, so the cycle
	// property is about code that must compile (review Fold 1).
	if cyclic := cycles(prodGraph); len(cyclic) > 0 {
		t.Fatalf("import cycles detected (0 expected): %v", cyclic)
	}
	properties++

	if properties != 3 {
		t.Fatalf("internal error: expected all 3 properties to run, got %d", properties)
	}

	path := filepath.Join(root, "tools", "arch", "baseline.txt")
	if *updateBaseline {
		writeBaseline(t, path, violations)
		t.Logf("baseline regenerated from the gate: %d violation(s)", len(violations))
		return
	}

	baseline := readBaseline(t, path)
	// An emptied baseline is a FAILURE only while violations exist (anti-bypass,
	// N-2). At the ratchet's terminal state — 0 violations — a well-formed
	// header-only baseline is the correct, green state (review F-1).
	if len(violations) > 0 && len(baseline) == 0 {
		t.Fatalf("baseline %s lists no violations but %d exist — an emptied baseline MUST fail", path, len(violations))
	}
	assertNoNewOrStale(t, path, violations, baseline)
}

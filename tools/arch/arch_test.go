//go:build arch

// The layer-discipline gate. Everything lives in this build-tagged test file so
// it runs only under `-tags=arch` (the Makefile `verify-architecture` target)
// and never in the default `go test ./...` run. See docs/decisions/0011-layer-
// discipline-gate.md (RULE-A/B/C/D) and docs/decisions/0016-application-import-
// ceiling.md (RULE-E) and specs/truth/techstack.md for the rules and policy.
package arch

import (
	"bytes"
	"flag"
	"go/ast"
	"go/parser"
	"go/token"
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
// is the PRIMARY owner for `make`-launched invocations; this filter keeps THIS
// gate's VERDICT hermetic on the gate's documented DIRECT invocation
// (`go test -count=1 -tags=arch … ./tools/arch`), which bypasses `make`. It does
// NOT make that path's outer `go test`/`go vet` hermetic (ADR 0012 R1). The two
// sites neutralise by different mechanisms (the Makefile block disables the env
// file + unsets; this filter re-sets explicit values), so the relation is
// COVERAGE (every name the block neutralises is re-set here or recorded as a
// known non-covered name — e.g. GOARM/GOEXPERIMENT under the frozen round-042
// guard, R4), NOT set-equality.
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

// sanctionedImports is the normative RULE-E allow-list (ADR 0016 D1/D3): for a
// governed application tier (internal/app/**, internal/cli) an import of an
// internal/** package is legal ONLY if it matches one of these entries. A
// trailing "/" marks a subtree entry; a bare name matches the exact package.
// This slice — the machine-readable source of the sanctioned set — is the single
// normative home for RULE-E; specs/truth/techstack.md cites it (ADR 0016), never
// restates it.
var sanctionedImports = []string{
	"internal/domain/", // tier 0 — the pure domain
	"internal/config",  // tier 1 — shared utility (leaf package)
	"internal/home",    // tier 1 — shared utility (leaf package)
	"internal/app/",    // tier 2 — application utilities (incl. the deps seam)
}

// matchesSanctioned reports whether package rel is covered by one sanctioned
// entry: a subtree entry (trailing "/") covers a prefix; a bare entry matches the
// exact package.
func matchesSanctioned(entry, rel string) bool {
	if strings.HasSuffix(entry, "/") {
		return strings.HasPrefix(rel, entry)
	}
	return rel == entry
}

// sanctioned reports whether a governed application tier may import rel under
// RULE-E. Default-deny: an internal/** package matching no entry is NOT allowed.
func sanctioned(rel string) bool {
	for _, entry := range sanctionedImports {
		if matchesSanctioned(entry, rel) {
			return true
		}
	}
	return false
}

// rel strips the module prefix; ok=false for stdlib / third-party imports.
func rel(importPath string) (string, bool) {
	if !strings.HasPrefix(importPath, modulePath) {
		return "", false
	}
	return strings.TrimPrefix(importPath, modulePath), true
}

// violation reports whether src -> dst breaks the layer predicate. RULE-A/B/C
// are ADR 0011 D2; RULE-E is ADR 0016 D1. Both tiers are already known-ranked.
func violation(src string, srcTier int, dst string, dstTier int) bool {
	switch {
	case dstTier > srcTier:
		return true // RULE-A (upward)
	case isApplicationTier(src) && strings.HasPrefix(dst, "internal/infrastructure/"):
		return true // RULE-B (application -> infrastructure)
	case strings.HasPrefix(src, "internal/domain/") && !strings.HasPrefix(dst, "internal/domain/"):
		return true // RULE-C (domain purity)
	case isApplicationTier(src) && !sanctioned(dst):
		return true // RULE-E (application import ceiling, default-deny)
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
	b.WriteString("# tools/arch/baseline.txt — layer-discipline gate baseline (ADR 0011 + ADR 0016).\n")
	b.WriteString("#\n")
	b.WriteString("# Generated from the gate's own output — NEVER hand-edit. Regenerate with:\n")
	b.WriteString("#   make verify-architecture-update\n")
	b.WriteString("#   (= go test -count=1 -tags=arch -run TestVerifyRealArchitecture ./tools/arch -args -update-baseline)\n")
	b.WriteString("#\n")
	b.WriteString("# Run the gate directly as:\n")
	b.WriteString("#   go vet -tags=arch ./tools/arch\n")
	b.WriteString("#   go test -count=1 -tags=arch -run TestVerifyRealArchitecture ./tools/arch\n")
	b.WriteString("#\n")
	b.WriteString("# A line `<src> -> <dst>` is a known, baselined layer violation: RULE-A/B/C/D\n")
	b.WriteString("# (ADR 0011) plus RULE-E, the application import ceiling (ADR 0016) — the\n")
	b.WriteString("# application tiers' residual unsanctioned edges. This is a ratchet that only\n")
	b.WriteString("# shrinks: a violation not listed here FAILS the gate, and a listed line that no\n")
	b.WriteString("# longer violates (stale) FAILS the gate too.\n")
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

// unusedSanctioned returns the sanctioned entries with no application-tier
// importer in graph. It reads the **merged** (production + test) graph — the same
// graph RULE-E governs (ADR 0016 D4) — so a test-only application-tier use keeps
// a sanctioned entry alive; that is deliberate, not incidental (review N-3).
func unusedSanctioned(graph map[string]map[string]bool) []string {
	used := map[string]bool{}
	for src, dsts := range graph {
		if !isApplicationTier(src) {
			continue
		}
		for dst := range dsts {
			if isInternal(dst) {
				used[dst] = true
			}
		}
	}
	var unused []string
	for _, entry := range sanctionedImports {
		hit := false
		for dst := range used {
			if matchesSanctioned(entry, dst) {
				hit = true
				break
			}
		}
		if !hit {
			unused = append(unused, entry)
		}
	}
	sort.Strings(unused)
	return unused
}

// assertSanctionedInUse asserts the fail-on-stale allow-list (ADR 0016 D4): every
// sanctioned entry must be imported by at least one governed application-tier
// package. An unused sanctioned entry FAILS — the allow-list, like the baseline,
// must shrink to truth (symmetry with ADR 0011 D3).
func assertSanctionedInUse(t *testing.T, graph map[string]map[string]bool) {
	t.Helper()
	if u := unusedSanctioned(graph); len(u) > 0 {
		t.Fatalf("sanctioned allow-list entr(ies) unused by any application-tier import (fail-on-stale allow-list, ADR 0016 D4): %v — remove them from the sanctioned set", u)
	}
}

// selfTestAllowList unit-tests the coverage predicate `unusedSanctioned` on
// synthetic graphs (review F-2): an all-used sanctioned set reports nothing; a
// set with one unimported entry reports exactly it. Without this witness the
// coverage assertion's own logic had no committed regression carrier (mutant M9
// escaped), which is the same class round 046's fold review flagged.
func selfTestAllowList(t *testing.T) {
	t.Helper()
	allUsed := map[string]map[string]bool{
		"internal/app/deps": {"internal/config": true},
		"internal/cli":      {"internal/domain/llm": true, "internal/home": true, "internal/app/deps": true},
	}
	if u := unusedSanctioned(allUsed); len(u) != 0 {
		t.Fatalf("allow-list self-test: an all-used sanctioned set reported stale entries: %v", u)
	}
	missingOne := map[string]map[string]bool{
		// internal/home is never imported by an application tier here.
		"internal/cli": {"internal/domain/llm": true, "internal/config": true, "internal/app/deps": true},
	}
	if u := unusedSanctioned(missingOne); len(u) != 1 || u[0] != "internal/home" {
		t.Fatalf("allow-list self-test: expected exactly [internal/home] unused, got %v", u)
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

// selfTestPredicate unit-tests the layer predicate on a synthetic graph,
// covering RULE-A/B/C/D (ADR 0011 D2) and RULE-E (ADR 0016 D1) plus the legal
// boundaries (sanctioned application-tier imports, downward imports).
func selfTestPredicate(t *testing.T) {
	t.Helper()
	graph := map[string]map[string]bool{
		"internal/agent":           {"internal/ui": true},
		"internal/app/deps":        {"internal/app/suggestions": true, "internal/config": true},
		"internal/app/suggestions": {"internal/infrastructure/tools": true, "internal/config": true},
		"internal/cli": {"internal/infrastructure/history": true, "internal/agent": true,
			"internal/ui": true, "internal/app/deps": true, "internal/domain/llm": true},
		"internal/domain/history":   {"internal/config": true},
		"internal/domain/llm":       {"internal/domain/history": true},
		"internal/home":             {"internal/domain/metrics": true},
		"internal/infrastructure/x": {"internal/config": true},
		"internal/infrastructure/z": {"internal/weird": true},
		"internal/ui/tui/prompt":    {"internal/domain/llm": true},
		"internal/weird":            {"internal/domain/llm": true},
	}
	want := []string{
		"internal/agent -> internal/ui",                             // RULE-A
		"internal/app/suggestions -> internal/infrastructure/tools", // RULE-A (2 -> 3, upward; also B/E)
		"internal/cli -> internal/agent",                            // RULE-E
		"internal/cli -> internal/infrastructure/history",           // RULE-B (6 -> 3, downward)
		"internal/cli -> internal/ui",                               // RULE-E
		"internal/domain/history -> internal/config",                // RULE-C
		"internal/weird -> (unranked governed package)",             // RULE-D
	}
	if got := evaluate(graph); !reflect.DeepEqual(got, want) {
		t.Fatalf("predicate self-test failed:\n got  %v\n want %v", got, want)
	}
	// RULE-E sanctioned boundaries: legal application-tier imports.
	if !sanctioned("internal/domain/llm") || !sanctioned("internal/config") ||
		!sanctioned("internal/home") || !sanctioned("internal/app/deps") {
		t.Fatalf("RULE-E sanctioned set self-test failed: a sanctioned package was rejected")
	}
	for _, unsanctioned := range []string{"internal/agent", "internal/ui", "internal/ui/tui/prompt", "internal/infrastructure/mcp"} {
		if sanctioned(unsanctioned) {
			t.Fatalf("RULE-E sanctioned set self-test failed: %q must NOT be sanctioned (default-deny)", unsanctioned)
		}
	}
}

// couplingSurface is the RULE-F allow-list (round 049 TD-1): the APPLICATION
// COUPLING SURFACE. RULE-E's granularity is the package EDGE; it is structurally
// blind to HOW MANY identifiers cross an already-baselined edge, so a shrunk edge
// can silently re-inflate (a second construction site, a new `agent.*` helper, a
// package-level `var _ agent.X`) with 0 new / 0 stale. RULE-F closes that gap: for
// a governed application tier named here, the set of IDENTIFIERS it may select
// from the listed target package is a normative allow-list (fail-on-stale).
//
// Key: "<src> -> <dst>" (module-relative). Value: the allowed exported-identifier
// set the src PRODUCTION sources may select from dst. The machine-readable home
// of this list is this table (like the sanctioned set, ADR 0016 D1).
//
// Metric (round-049 fold review N-8): RULE-F counts DISTINCT package-qualified
// SELECTORS (`alias.Ident`) — which named symbols cross the boundary — NOT call
// sites, NOT method calls on type values, NOT struct fields. So `internal/cli`'s
// `→ ui` surface is 19 identifiers while the edge carries ~29 `ui.X` occurrences
// and ~4 method calls; the identifier count is deliberately NOT a call-site
// measure (ADR 0017 §Forward scopes the `→ ui` slice by call sites).
//
// Growth (round-049 fold review N-6): unlike RULE-E's baseline, this table has NO
// regeneration affordance — `make verify-architecture-update` rewrites
// baseline.txt and never touches couplingSurface. Intentional growth (a new
// tracked identifier) is therefore a guard-TABLE edit in the same PR.
var couplingSurface = map[string]map[string]bool{
	"internal/cli -> internal/agent": {"AgentLoop": true},
	"internal/cli -> internal/ui": {
		"ComputeCost":              true,
		"DefaultToolOutputIdleGap": true,
		"FormatInputCaptured":      true,
		"FormatMetrics":            true,
		"FormatPayloadStatus":      true,
		"FormatReady":              true,
		"FormatToolReason":         true,
		"FormatToolUsage":          true,
		"FormatTurnGap":            true,
		"FormatTurnOpening":        true,
		"HitRate":                  true,
		"NewRenderer":              true,
		"NewSpinner":               true,
		"NewToolOutputCoordinator": true,
		"Pricing":                  true,
		"Spinner":                  true,
		"ToolLineRenderer":         true,
		"ToolUsageRow":             true,
		"UsageCounts":              true,
	},
}

// productionGoFiles returns the non-_test.go file paths directly under dir.
func productionGoFiles(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("cannot read %s: %v", dir, err)
	}
	var out []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		out = append(out, filepath.Join(dir, name))
	}
	return out
}

// scanCouplingSurface parses every governed application-tier PRODUCTION file under
// root and returns, per coupling-surface edge, the set of identifiers selected
// from the target package. Build-tag-gated application files are evaluated as
// written (the union over tags is not simulated; the application tiers carry no
// tag-gated files today — a future addition must widen this, ADR 0011 D6).
func scanCouplingSurface(t *testing.T, root string) map[string]map[string]bool {
	t.Helper()
	got := map[string]map[string]bool{}
	srcs := map[string]bool{}
	for edge := range couplingSurface {
		got[edge] = map[string]bool{}
		src, _, _ := strings.Cut(edge, " -> ")
		srcs[src] = true
	}
	for src := range srcs {
		dir := filepath.Join(root, filepath.FromSlash(src))
		for _, path := range productionGoFiles(t, dir) {
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, path, nil, 0)
			if err != nil {
				t.Fatalf("cannot parse %s: %v", path, err)
			}
			// Resolve each edge-target's local alias in THIS file.
			alias := map[string]string{} // alias -> edge key
			for _, spec := range f.Imports {
				rd, ok := rel(strings.Trim(spec.Path.Value, `"`))
				if !ok {
					continue
				}
				edge := src + " -> " + rd
				if _, tracked := couplingSurface[edge]; !tracked {
					continue
				}
				local := rd[strings.LastIndex(rd, "/")+1:]
				if spec.Name != nil {
					local = spec.Name.Name
				}
				if local == "_" {
					continue
				}
				if local == "." {
					t.Fatalf("%s uses a dot-import of %s — RULE-F cannot attribute bare selectors; rename or alias the import", path, rd)
				}
				alias[local] = edge
			}
			if len(alias) == 0 {
				continue
			}
			ast.Inspect(f, func(n ast.Node) bool {
				sel, ok := n.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				x, ok := sel.X.(*ast.Ident)
				if !ok {
					return true
				}
				if edge, ok := alias[x.Name]; ok {
					got[edge][sel.Sel.Name] = true
				}
				return true
			})
		}
	}
	return got
}

// diffSurface returns the new (got∖allowed) and stale (allowed∖got) identifiers.
func diffSurface(allowed, got map[string]bool) (fresh, stale []string) {
	for id := range got {
		if !allowed[id] {
			fresh = append(fresh, id)
		}
	}
	for id := range allowed {
		if !got[id] {
			stale = append(stale, id)
		}
	}
	sort.Strings(fresh)
	sort.Strings(stale)
	return fresh, stale
}

// assertCouplingSurface applies RULE-F: every tracked edge's identifier set must
// equal its allow-list — a new identifier FAILS (a re-inflated coupling surface),
// and an allow-list entry no longer selected FAILS (stale, the ratchet shrinks to
// truth).
func assertCouplingSurface(t *testing.T, root string) {
	t.Helper()
	got := scanCouplingSurface(t, root)
	for _, edge := range slices.Sorted(maps.Keys(couplingSurface)) {
		fresh, stale := diffSurface(couplingSurface[edge], got[edge])
		if len(fresh) > 0 {
			t.Errorf("RULE-F: %s — identifier(s) not in the coupling-surface allow-list: %v (a tracked edge's surface grew; if intentional, extend couplingSurface in the same PR — the edge ratchet's -update-baseline does not apply here)", edge, fresh)
		}
		if len(stale) > 0 {
			t.Errorf("RULE-F: %s — STALE allow-list entr(ies) no longer referenced: %v (remove them from couplingSurface)", edge, stale)
		}
	}
}

// selfTestCouplingSurface unit-tests the RULE-F diff predicate on synthetic sets
// (review F-2 precedent): an exact match reports nothing; an extra identifier is
// "new"; a dropped allow-list entry is "stale".
func selfTestCouplingSurface(t *testing.T) {
	t.Helper()
	if f, s := diffSurface(map[string]bool{"AgentLoop": true}, map[string]bool{"AgentLoop": true}); len(f) != 0 || len(s) != 0 {
		t.Fatalf("surface self-test: exact match reported fresh=%v stale=%v", f, s)
	}
	if f, s := diffSurface(map[string]bool{"AgentLoop": true}, map[string]bool{"AgentLoop": true, "ToolDefs": true}); len(f) != 1 || f[0] != "ToolDefs" || len(s) != 0 {
		t.Fatalf("surface self-test: expected fresh=[ToolDefs], got fresh=%v stale=%v", f, s)
	}
	if f, s := diffSurface(map[string]bool{"AgentLoop": true, "ToolDefs": true}, map[string]bool{"AgentLoop": true}); len(f) != 0 || len(s) != 1 || s[0] != "ToolDefs" {
		t.Fatalf("surface self-test: expected stale=[ToolDefs], got fresh=%v stale=%v", f, s)
	}
}

// assertSurfaceCoversBaseline (round-049 fold review TD-2) asserts RULE-F's
// coverage invariant over the **governed application edges** — the application-
// tier edges the gate currently SEES (the computed violation set, which subsumes
// the baseline and, unlike reading the baseline file, also fires on a brand-new
// not-yet-baselined edge). Every such edge MUST have a couplingSurface entry.
// Without it, RULE-F's protection is opt-in by memory: a new application edge,
// once absorbed into the baseline via the documented `-update-baseline`
// regeneration path, would be governed by the edge ratchet but have an
// unprotected identifier surface. This is the direct analogue of RULE-E's
// assertSanctionedInUse (ADR 0016 D4).
func assertSurfaceCoversBaseline(t *testing.T, governedAppEdges []string) {
	t.Helper()
	var uncovered []string
	for _, line := range governedAppEdges {
		src, dst, ok := strings.Cut(line, " -> ")
		if !ok || !isApplicationTier(src) || strings.Contains(dst, "(unranked") {
			continue
		}
		if _, tracked := couplingSurface[line]; !tracked {
			uncovered = append(uncovered, line)
		}
	}
	if len(uncovered) > 0 {
		sort.Strings(uncovered)
		t.Errorf("RULE-F coverage: governed application edge(s) with no couplingSurface entry: %v — governed by the edge ratchet but with an unprotected identifier surface (add a couplingSurface entry)", uncovered)
	}
}

// TestVerifyRealArchitecture is the gate. It runs all four properties —
// enumeration, ranking/baseline diff, acyclicity, coupling surface — and asserts they all ran, so
// `-run TestVerifyRealArchitecture` alone cannot silently skip one (N-3).
func TestVerifyRealArchitecture(t *testing.T) {
	root := moduleRoot(t)

	properties := 0

	// Property 1 — enumeration (B-2).
	graph, prodGraph := enumerate(t, root)
	assertGraphEnumerated(t, graph)
	violations := evaluate(graph)
	properties++

	// The self-tests: two synthetic predicates (the layer predicate, the
	// coupling-surface diff predicate) plus four real-graph assertions (RULE-D
	// tier coverage, RULE-E allow-list-in-use, RULE-F surface, RULE-F coverage).
	// Each runs as a named subtest and the executed
	// NAME SET is asserted, so a mutant that drops a call (or an edit that renames
	// one) cannot slip through a hand-maintained counter the way M7 did — and a
	// deleted call reds rather than silently passing (review N-2 / F-2).
	//
	// Ordering is load-bearing: these run BEFORE the `*updateBaseline` branch
	// below, so `make verify-architecture-update` cannot launder a stale allow-list
	// entry into a freshly generated baseline (review §1). Keep them ahead of it.
	// The name set is compared ORDER-SENSITIVELY (reflect.DeepEqual) by design:
	// the load-bearing property is "all seven run before the *updateBaseline
	// branch", and the order is also pinned (review N-3′).
	wantSelfTests := []string{"predicate", "allow-list", "tier-coverage", "sanctioned-in-use", "surface-predicate", "surface", "surface-coverage"}
	var ranSelfTests []string
	runSelfTest := func(name string, fn func(*testing.T)) {
		// Record the ATTEMPT, not the pass: run the subtest then append its name
		// unconditionally, so a genuine subtest failure reports itself at the
		// subtest rather than surfacing as a misleading "self-test did not run"
		// internal error (review N-1′).
		t.Run(name, fn)
		ranSelfTests = append(ranSelfTests, name)
	}
	runSelfTest("predicate", selfTestPredicate)                                              // RULE-A/B/C/D + RULE-E, synthetic (ADR 0011 D2 / ADR 0016 D1)
	runSelfTest("allow-list", selfTestAllowList)                                             // RULE-E coverage predicate, synthetic (ADR 0016 D4 / review F-2)
	runSelfTest("tier-coverage", func(t *testing.T) { assertNoUnrankedGoverned(t, graph) })  // RULE-D coverage, real graph
	runSelfTest("sanctioned-in-use", func(t *testing.T) { assertSanctionedInUse(t, graph) }) // RULE-E allow-list in use, real graph (ADR 0016 D4)
	runSelfTest("surface-predicate", func(t *testing.T) { selfTestCouplingSurface(t) })      // RULE-F diff predicate, synthetic (round 049 TD-1)

	// Property 4 — RULE-F: the application coupling surface. The baseline ratchet
	// governs EDGES; RULE-F governs the IDENTIFIERS crossing a tracked edge, so a
	// shrunk edge cannot silently re-inflate (round 049 TD-1). Wired as NAMED
	// self-tests (not a bare call) so the name-set defence above protects them from
	// a silent drop (review F-2), and they run BEFORE the `*updateBaseline` branch
	// (regenerating the baseline must not launder a re-inflated surface green).
	runSelfTest("surface", func(t *testing.T) { assertCouplingSurface(t, root) })                      // RULE-F identifier allow-list, real graph (round 049 TD-1)
	runSelfTest("surface-coverage", func(t *testing.T) { assertSurfaceCoversBaseline(t, violations) }) // RULE-F coverage of baselined app edges, real graph (round 049 TD-2)
	if !reflect.DeepEqual(ranSelfTests, wantSelfTests) {
		t.Fatalf("internal error: expected self-tests %v to run, got %v", wantSelfTests, ranSelfTests)
	}

	// Property 2 — ranking + baseline diff (rule evaluated on the merged graph,
	// so test imports are governed; `violations` is computed in Property 1).
	properties++

	// Property 3 — acyclicity (ADR 0011 D8 / issue #93 AC4), on the
	// **production-only** union: Go permits test-only cycles, so the cycle
	// property is about code that must compile (review Fold 1).
	if cyclic := cycles(prodGraph); len(cyclic) > 0 {
		t.Fatalf("import cycles detected (0 expected): %v", cyclic)
	}
	properties++

	// Property 4 — RULE-F (the coupling surface + its coverage of baselined
	// application edges) ran as the named self-tests `surface` / `surface-coverage`
	// above; count the property here so a deletion of the whole block is caught.
	properties++

	if properties != 4 {
		t.Fatalf("internal error: expected all 4 properties to run, got %d", properties)
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

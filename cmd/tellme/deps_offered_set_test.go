package main

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// Round 090 (issue #186) — the offered-set doc-consistency carrier.
//
// The truth surface `chat/dsl.md` documents the offered agent-tool set for the
// `the request offered exactly the agent tools` step. That documentation must be
// **bound to** the live production registry so it cannot silently rot again (the
// round-088 F-088-1 / round-089 RF-089-6 class: a docs-prose claim with no
// mechanical carrier). This test binds the doc's `集合` cell to `agentTools()` —
// the NON-overridable production assembler — so adding/removing a tool, or
// editing the documented list, REDDENS here instead of drifting undetected.
// (It enforces doc↔registry *consistency*, not the *size* of the surface.)
//
// It is a UNIT test (rides `go test` / `make test`), not a `make verify` member —
// the same tier as the round-031 schema gate and the round-061 projection gate.

// offeredSetDocPath is the truth surface that enumerates the offered set,
// relative to the cmd/tellme package dir (where `go test` runs).
const offeredSetDocPath = "../../specs/truth/features/cli/chat/dsl.md"

// offeredSetRowMarker identifies the DSL row whose `集合` cell enumerates the set.
const offeredSetRowMarker = "the request offered exactly the agent tools"

// backtickedName matches a backticked wire tool name (`list_files`, …).
var backtickedName = regexp.MustCompile("`([a-z_][a-z0-9_]*)`")

// documentedOfferedSet parses the `集合` cell of the offered-set DSL row and
// returns its backticked tool names, sorted.
func documentedOfferedSet(t *testing.T) []string {
	t.Helper()
	data, err := os.ReadFile(offeredSetDocPath)
	if err != nil {
		t.Fatalf("read %s: %v", offeredSetDocPath, err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "|") || !strings.Contains(line, offeredSetRowMarker) {
			continue
		}
		// The row is a pipe-delimited Markdown table line; the `集合` cell is
		// the structured enumeration (its sibling cells carry prose/round notes).
		for _, cell := range strings.Split(line, "|") {
			if !strings.Contains(cell, "`集合`") {
				continue
			}
			var names []string
			for _, m := range backtickedName.FindAllStringSubmatch(cell, -1) {
				names = append(names, m[1])
			}
			if len(names) == 0 {
				t.Fatalf("the `集合` cell of the offered-set row enumerates no tool names: %q", strings.TrimSpace(cell))
			}
			sort.Strings(names)
			return names
		}
		t.Fatalf("the offered-set row has no `集合` cell: %q", strings.TrimSpace(line))
	}
	t.Fatalf("%s carries no row for the offered-set step %q", offeredSetDocPath, offeredSetRowMarker)
	return nil
}

// TestOfferedSetDocMatchesTheLiveRegistry is the round-090 carrier: the set the
// truth doc enumerates MUST equal the live base agent-tool set (`agentTools()`).
func TestOfferedSetDocMatchesTheLiveRegistry(t *testing.T) {
	doc := documentedOfferedSet(t)

	live := make([]string, 0, len(agentTools()))
	for _, tl := range agentTools() {
		live = append(live, tl.Name())
	}
	sort.Strings(live)

	if len(doc) == 0 || len(live) == 0 {
		t.Fatal("the carrier would pass vacuously — the doc or the registry enumerates nothing")
	}
	if strings.Join(doc, ",") != strings.Join(live, ",") {
		t.Fatalf("the documented offered set is not single-sourced to the live registry:\n  doc  (%d): %v\n  live (%d): %v\n"+
			"reconcile %s (the `集合` cell) with cmd/tellme agentTools()", len(doc), doc, len(live), live, offeredSetDocPath)
	}
}

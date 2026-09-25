package main

import (
	"strings"
	"testing"

	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
	infratools "github.com/gosharplite/tellme/internal/infrastructure/tools"
)

// TestAgentToolsIsTheCanonicalBaseSet pins round 092 (issue #189; ADR 0062): the
// production BASE set — `agentTools()` (the non-overridable assembler the
// round-031 schema gate iterates) — MUST equal the single canonical owner
// `infratools.NewAgentBaseTools(nil)`, which `assembleAgentTools` delegates to.
// If a future edit re-inlines the base composition in `assembleAgentTools` and it
// diverges from the owner, this reddens (the round-090 equality-of-surfaces
// carrier shape). It is a UNIT test (rides `go test` / `make test`), not a
// `make verify` member — the same tier as the round-090 offered-set carrier.
func TestAgentToolsIsTheCanonicalBaseSet(t *testing.T) {
	namesOf := func(ts []domaintools.Tool) []string {
		out := make([]string, 0, len(ts))
		for _, tl := range ts {
			out = append(out, tl.Name())
		}
		return out
	}
	owner := namesOf(infratools.NewAgentBaseTools(nil))
	got := namesOf(agentTools())
	if len(owner) == 0 {
		t.Fatal("the canonical base-set owner enumerates no tools — the carrier would pass vacuously")
	}
	if strings.Join(owner, ",") != strings.Join(got, ",") {
		t.Fatalf("agentTools() is not the canonical base set:\n  owner (%d): %v\n  agent (%d): %v\n"+
			"delegate assembleAgentTools' base half to infratools.NewAgentBaseTools", len(owner), owner, len(got), got)
	}
}

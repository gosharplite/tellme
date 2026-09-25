package steps

import (
	"strings"
	"testing"

	infratools "github.com/gosharplite/tellme/internal/infrastructure/tools"
)

// TestRegisteredToolNamesIsTheCanonicalBaseSet pins round 092 (issue #189;
// ADR 0062): the e2e enumerator `registeredToolNames()` MUST equal the single
// canonical owner `infratools.NewAgentBaseTools(nil)`, which it delegates to. If a
// future edit re-inlines the base composition in `registeredToolNames` and it
// diverges from the owner, this reddens — so the harness can no longer be edited
// out of step with the production registry (the round-090 equality-of-surfaces
// carrier shape).
func TestRegisteredToolNamesIsTheCanonicalBaseSet(t *testing.T) {
	owner := infratools.NewAgentBaseTools(nil)
	want := make([]string, 0, len(owner))
	for _, tl := range owner {
		want = append(want, tl.Name())
	}
	got := registeredToolNames()
	if len(want) == 0 {
		t.Fatal("the canonical base-set owner enumerates no tools — the carrier would pass vacuously")
	}
	if strings.Join(want, ",") != strings.Join(got, ",") {
		t.Fatalf("registeredToolNames() is not the canonical base set:\n  owner (%d): %v\n  enumerator (%d): %v\n"+
			"delegate registeredToolNames to infratools.NewAgentBaseTools", len(want), want, len(got), got)
	}
}

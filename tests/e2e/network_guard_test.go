package e2e

import (
	"testing"

	"github.com/gosharplite/tellme/tests/e2e/harness"
)

// assertNoNetworkCapability is the build-graph capability guard: it fails if the
// built binary can perform network I/O (a network-capable closure package, or a
// dialing/listening symbol in the linked binary).
func assertNoNetworkCapability(t *testing.T) {
	t.Helper()
	violation, err := harness.NetworkCapabilityViolation()
	if err != nil {
		t.Fatalf("network capability check: %v", err)
	}
	if violation != "" {
		t.Errorf("network capability detected in ./cmd/tellme: %s", violation)
	}
}

// TestDependencyGraphHasNoNetworkCapability proves the capability-absence witness
// and smokes the differential harness wiring. The child environment is owned by
// the test (TELL_ME_* unset) so the outcome does not depend on the ambient shell.
func TestDependencyGraphHasNoNetworkCapability(t *testing.T) {
	assertNoNetworkCapability(t)

	unset := []string{"TELL_ME_HOME", "TELL_ME_MODE", "TELL_ME_SELECTED_PROVIDER"}
	base := harness.Run(nil, nil, unset)
	blocked := harness.RunWithBlockedNetwork(nil, nil, unset)
	if base.Err != nil || blocked.Err != nil {
		t.Fatalf("run error: base=%v blocked=%v", base.Err, blocked.Err)
	}
	if base.ExitCode != blocked.ExitCode || base.Stdout != blocked.Stdout {
		t.Errorf("hostile network changed the outcome: base(exit=%d out=%q) blocked(exit=%d out=%q)",
			base.ExitCode, base.Stdout, blocked.ExitCode, blocked.Stdout)
	}
}

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

// hostileNetworkEnv returns environment overrides that make any real egress fail
// fast without privileges (an unroutable HTTP(S) proxy). This is the portable
// no-egress witness for SC-004 — the privileged netns is not available on the
// local dev host.
func hostileNetworkEnv() map[string]string {
	return map[string]string{
		"HTTP_PROXY":  "http://127.0.0.1:1",
		"HTTPS_PROXY": "http://127.0.0.1:1",
		"NO_PROXY":    "",
		"http_proxy":  "http://127.0.0.1:1",
		"https_proxy": "http://127.0.0.1:1",
		"no_proxy":    "",
	}
}

// runWithBlockedNetwork runs the built tellme with args under a hostile network
// environment. A caller compares the result to the same run under a normal
// environment to obtain the differential no-egress witness.
func runWithBlockedNetwork(args []string, env map[string]string) harness.RunResult {
	merged := make(map[string]string, len(env)+len(hostileNetworkEnv()))
	for k, v := range env {
		merged[k] = v
	}
	for k, v := range hostileNetworkEnv() {
		merged[k] = v
	}
	return harness.Run(args, merged, nil)
}

// TestDependencyGraphHasNoNetworkCapability proves the capability-absence
// witness and smokes the differential harness wiring.
func TestDependencyGraphHasNoNetworkCapability(t *testing.T) {
	assertNoNetworkCapability(t)

	base := harness.Run(nil, nil, nil)
	blocked := runWithBlockedNetwork(nil, nil)
	if base.Err != nil || blocked.Err != nil {
		t.Fatalf("run error: base=%v blocked=%v", base.Err, blocked.Err)
	}
	if base.ExitCode != blocked.ExitCode || base.Stdout != blocked.Stdout {
		t.Errorf("hostile network changed the outcome: base(exit=%d out=%q) blocked(exit=%d out=%q)",
			base.ExitCode, base.Stdout, blocked.ExitCode, blocked.Stdout)
	}
}

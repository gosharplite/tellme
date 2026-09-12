package harness

// HostileNetworkEnv returns environment overrides that make any real egress fail
// fast without privileges (an unroutable HTTP(S) proxy). It is the portable
// no-egress witness for the offline paths — the privileged netns is not
// available on the local dev host. This is the single definition of the
// differential wiring.
func HostileNetworkEnv() map[string]string {
	return map[string]string{
		"HTTP_PROXY":  "http://127.0.0.1:1",
		"HTTPS_PROXY": "http://127.0.0.1:1",
		"NO_PROXY":    "",
		"http_proxy":  "http://127.0.0.1:1",
		"https_proxy": "http://127.0.0.1:1",
		"no_proxy":    "",
	}
}

// RunWithBlockedNetwork runs the built binary with args under a hostile network
// environment: `set` overrides plus the hostile overrides, with `unset` names
// removed. A caller compares the result to a normal run to obtain the
// differential no-egress witness.
//
// Round 004 retired the whole-binary build-graph capability guard (the
// prompt-bearing chat path legitimately links net/http); the offline guarantee
// is now an offline-path behaviour claim — a recording sink left untouched plus
// this differential witness.
func RunWithBlockedNetwork(args []string, set map[string]string, unset []string) RunResult {
	merged := make(map[string]string, len(set)+len(HostileNetworkEnv()))
	for k, v := range set {
		merged[k] = v
	}
	for k, v := range HostileNetworkEnv() {
		merged[k] = v
	}
	return Run(args, merged, unset)
}

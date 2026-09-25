package harness

import "testing"

// TestSyncedStdinGateWaitsForTheComposedText pins the round-093 gate RULE: the
// paint gate that releases the terminal key is CONTENT-AWARE — the composed
// text's last visible line, not a generic marker any (possibly partial) frame
// satisfies. Gutting paintGate (returning the fallback) reddens this pin.
//
// Coverage scope (architect review F-093-1): this pin exercises the RULE, not
// its call site. Reverting runExecSynced's call site
// (`marker := paintGate(compose, fallbackMarker)` → `marker := fallbackMarker`)
// re-opens the round-191 flake yet leaves this pin green — the wiring is covered
// only by the (probabilistic) E2E repetition witness. "Gut paintGate" and
// "revert the call site" are NOT equivalent mutations. A deterministic
// ordering/wiring carrier is the forward option ADR 0063 §Forward RF-093-1.
//
// The pre-fix flake magnitude is host/load-specific (F-093-2): ~62 % red at
// 16-way concurrency was measured on the authoring host (linux/amd64); the
// review host (darwin/amd64) measured ~2.8 %. The direction (0 fixed vs > 0
// pre-fix) reproduces; the size does not.
func TestSyncedStdinGateWaitsForTheComposedText(t *testing.T) {
	const border = "┌"
	cases := []struct {
		name    string
		compose string
		want    string
	}{
		{"multi-line compose gates on the last visible line", "line one\rline two", "line two"},
		{"single-line compose gates on the whole line", "deploy to staging", "deploy to staging"},
		{"a leading control key is not visible text", "\x13hi", "hi"},
		{"a trailing control key is not visible text", "hi\x03", "hi"},
		{"an open/abort-only sequence falls back to the border", "", border},
		{"a control-only sequence falls back to the border", "\x03", border},
	}
	for _, tc := range cases {
		if got := paintGate(tc.compose, border); got != tc.want {
			t.Errorf("%s: paintGate(%q) = %q, want %q", tc.name, tc.compose, got, tc.want)
		}
	}
}

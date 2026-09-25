package harness

import "testing"

// TestSyncedStdinGateWaitsForTheComposedText pins the round-093 ordering rule:
// the paint gate that releases the terminal key is CONTENT-AWARE — it is the
// composed text's last visible line, not a generic marker any (possibly partial)
// frame satisfies. A revert to a constant gate (e.g. the editor border) reddens
// this pin and re-opens the round-191 flake (measured ~63 % red at 16-way
// concurrency).
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

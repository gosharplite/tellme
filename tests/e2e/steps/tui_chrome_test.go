package steps

import "testing"

// TestTuiCursorRowsCountsOnlyCursorRows pins the suggestion-cursor predicate: it
// must count only the rendered cursor row (the exact `"  > "` prefix) and must NOT
// count a suggestion whose TEXT itself begins with `> ` (an unselected row renders
// with the 4-space row prefix). Round-037 review G-2: the earlier `TrimLeft` proxy
// miscounted such a suggestion and produced a false failure once the predicate
// became an exclusion.
func TestTuiCursorRowsCountsOnlyCursorRows(t *testing.T) {
	cases := []struct {
		name string
		out  string
		want int
	}{
		{"cursor row is counted", "  Suggestions:\n  > deploy to staging\n    review\n", 1},
		{"no rows at rest", "  Suggestions:\n    deploy to staging\n    review the last commit\n", 0},
		{"a suggestion whose text starts with '> ' is not a cursor row",
			"  Suggestions:\n    > quoted reply\n    deploy to staging\n", 0},
		{"empty output", "", 0},
	}
	for _, tc := range cases {
		if got := tuiCursorRows(tc.out); got != tc.want {
			t.Errorf("%s: tuiCursorRows = %d, want %d\nout=%q", tc.name, got, tc.want, tc.out)
		}
	}
}

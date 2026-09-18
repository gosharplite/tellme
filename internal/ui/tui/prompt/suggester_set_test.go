package prompt

import (
	"strings"
	"testing"
)

// Round 052 (closes #115 R-1; ADR 0021): the selection policy is owned by the
// CALLER — `set(items, cursor)` takes the cursor explicitly and no longer resets
// it internally. These pins witness that; reverting to an internal reset reds the
// "explicit cursor selects that row" case.

func TestSuggesterSetOwnsTheCursor(t *testing.T) {
	items := []string{"alpha", "beta", "gamma"}

	t.Run("noChoice leaves nothing selected", func(t *testing.T) {
		s := newSuggester()
		s.set(items, noChoice)
		if s.cursor != noChoice {
			t.Fatalf("cursor = %d, want noChoice (%d)", s.cursor, noChoice)
		}
		if got := s.selected(); got != "" {
			t.Fatalf("selected() = %q, want none", got)
		}
	})

	t.Run("an explicit cursor selects that row", func(t *testing.T) {
		s := newSuggester()
		s.set(items, 0)
		if got := s.selected(); got != "alpha" {
			t.Fatalf("selected() = %q, want alpha", got)
		}
	})

	t.Run("set replaces the items and the cursor together", func(t *testing.T) {
		s := newSuggester()
		s.set([]string{"alpha", "beta", "gamma"}, 0)
		s.set([]string{"delta"}, noChoice)
		if got := s.selected(); got != "" {
			t.Fatalf("selected() after a refresh = %q, want none", got)
		}
	})

	t.Run("cycle from noChoice lands on the first row", func(t *testing.T) {
		s := newSuggester()
		s.set(items, noChoice)
		s.cycle(+1)
		if got := s.selected(); got != "alpha" {
			t.Fatalf("selected() after cycle(+1) = %q, want alpha", got)
		}
	})

	// RF-52-3: `set` trusts its caller, but `selected()` is total and `view()` is
	// safe on an out-of-range cursor — no panic, no stray highlight. The input is
	// NON-EMPTY (round-052 fold F-52-3): with an empty list `len(items) == 0`
	// short-circuits before the cursor is consulted, so it would exercise the
	// empty-list path, not the upper bound. This case is the carrier for the
	// `cursor >= len(items)` half — deleting that bound reds it with
	// `index out of range [3] with length 1`.
	t.Run("an out-of-range cursor is total and safe", func(t *testing.T) {
		s := newSuggester()
		s.set([]string{"alpha"}, 3)
		if got := s.selected(); got != "" {
			t.Fatalf("selected() = %q, want none", got)
		}
		// With a non-empty list view() renders a header + rows; "no cursor row"
		// is the falsifiable assertion (an empty-string form would be vacuous).
		if v := s.view(); strings.Contains(v, "> ") {
			t.Fatalf("view() = %q, want NO cursor row", v)
		}
	})
}

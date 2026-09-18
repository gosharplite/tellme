package prompt

import "testing"

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
	// safe on an out-of-range cursor — no panic, no stray highlight. Pinned so a
	// future caller bug is a visible no-selection, not a crash.
	t.Run("an out-of-range cursor is total and safe", func(t *testing.T) {
		s := newSuggester()
		s.set(nil, 3)
		if got := s.selected(); got != "" {
			t.Fatalf("selected() = %q, want none", got)
		}
		if got := s.view(); got != "" {
			t.Fatalf("view() = %q, want empty", got)
		}
	})
}

package ui

import (
	"strings"
	"testing"
	"time"
)

// Round-025 unit tests (T006): the BOUNDED several-tool label and the row-aware
// clear (rows = ceil(width/columns), driven by an injected width seam).

func TestExecutingToolsLabelBounded(t *testing.T) {
	tests := []struct {
		names []string
		want  string
	}{
		{nil, " Executing tools..."},
		{[]string{}, " Executing tools..."},
		{[]string{"a"}, " Executing [a]..."},
		{[]string{"a", "b"}, " Executing tools [a and 1 more]..."},
		{[]string{"read_files", "list_files", "execute_command", "get_tree"}, " Executing tools [read_files and 3 more]..."},
	}
	for _, tt := range tests {
		if got := ExecutingToolsLabel(tt.names); got != tt.want {
			t.Errorf("ExecutingToolsLabel(%v) = %q, want %q", tt.names, got, tt.want)
		}
	}
	// The several-tool form must NOT enumerate the remaining names.
	got := ExecutingToolsLabel([]string{"alpha", "bravo", "charlie"})
	if strings.Contains(got, "bravo") || strings.Contains(got, "charlie") {
		t.Errorf("the several-tool label enumerated the remaining names: %q", got)
	}
}

func TestEraseRows(t *testing.T) {
	if got := eraseRows(0); got != clearControl {
		t.Errorf("eraseRows(0) = %q, want %q", got, clearControl)
	}
	if got := eraseRows(1); got != clearControl {
		t.Errorf("eraseRows(1) = %q, want %q", got, clearControl)
	}
	want := clearControl + cursorUp + clearControl + cursorUp + clearControl
	if got := eraseRows(3); got != want {
		t.Errorf("eraseRows(3) = %q, want %q", got, want)
	}
}

func TestRowsForLine(t *testing.T) {
	cases := []struct {
		line    string
		columns int
		want    int
	}{
		{"abc", 0, 1},          // unknown width → single-row best effort
		{"", 10, 1},            // empty
		{"abcdefghij", 10, 1},  // exactly the width
		{"abcdefghijk", 10, 2}, // one over
		{strings.Repeat("x", 73), 40, 2},
		{"⠋ Thinking [m]... (0s)", 40, 1},
	}
	for _, c := range cases {
		if got := rowsForLine(c.line, c.columns); got != c.want {
			t.Errorf("rowsForLine(%q, %d) = %d, want %d", c.line, c.columns, got, c.want)
		}
	}
}

// TestSpinnerRowAwareClear pins the round-025 row-aware clear: with an injected
// narrow width, the clear erases every row the wrapped frame occupied.
func TestSpinnerRowAwareClear(t *testing.T) {
	w := newSignalWriter()
	fixed := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	s := NewSpinner(w, "m", fixed, nil, func() int { return 40 })
	s.newTicker = func() (<-chan time.Time, func()) { return make(chan time.Time), func() {} }
	s.now = func() time.Time { return fixed }

	// A several-tool frame wider than 40 columns wraps across 2 rows.
	s.OnToolsStart([]string{"read_files", "list_files", "execute_command", "get_tree"})
	awaitWrite(t, w)
	s.Stop()
	if !strings.HasSuffix(w.String(), eraseRows(2)) {
		t.Errorf("Stop did not erase every wrapped row: %q", w.String())
	}
}

// TestSpinnerSingleRowClearUnchanged pins that a frame that fits one row keeps the
// round-019 single-row clear (`\r\x1b[K`).
func TestSpinnerSingleRowClearUnchanged(t *testing.T) {
	w := newSignalWriter()
	fixed := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	s := NewSpinner(w, "m", fixed, nil, func() int { return 40 })
	s.newTicker = func() (<-chan time.Time, func()) { return make(chan time.Time), func() {} }
	s.now = func() time.Time { return fixed }

	s.OnInferenceStart() // " Thinking [m]... (0s)" fits in 40 columns
	awaitWrite(t, w)
	s.Stop()
	if !strings.HasSuffix(w.String(), clearControl) {
		t.Errorf("a single-row frame's clear changed: %q", w.String())
	}
}

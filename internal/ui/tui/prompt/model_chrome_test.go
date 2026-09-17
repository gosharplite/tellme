package prompt

import (
	"context"
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// fakeSource is a canned suggestion source.
type fakeSource struct {
	items []string
}

func (f *fakeSource) Suggest(context.Context, string) []string { return f.items }

// maxLineLen returns the longest visible line length (go test renders with the
// Ascii profile, so lipgloss emits no ANSI here).
func maxLineLen(s string) int {
	max := 0
	for _, ln := range strings.Split(s, "\n") {
		if n := len([]rune(ln)); n > max {
			max = n
		}
	}
	return max
}

// TestModelRendersReferenceChrome (round-016 T017; round 037): the model renders the
// framed editor + styled suggestion list with NO session metrics header (strict parity),
// and — round 037 — NO suggestion pre-selected (the `>` cursor appears on ZERO rows at
// rest, aligning to the reference's `suggester{Index: -1}`). The exact cursor count is
// asserted on the single View() frame (round-016 implementation-review BLOCKER: the E2E
// capture accumulates frames, so exactness lives here).
func TestModelRendersReferenceChrome(t *testing.T) {
	m := New(context.Background(), strings.NewReader(""), &strings.Builder{}, &fakeSource{items: []string{"deploy to staging with version 016"}})
	view := m.View()
	if !strings.Contains(view, editorPlaceholder) {
		t.Fatalf("View() missing the placeholder: %q", view)
	}
	if !strings.Contains(view, "┌") {
		t.Fatalf("View() missing the editor border: %q", view)
	}
	if !strings.Contains(view, suggesterHeader) {
		t.Fatalf("View() missing the %q header: %q", suggesterHeader, view)
	}
	if n := strings.Count(view, "> "); n != 0 {
		t.Fatalf("View() selection cursor count = %d, want exactly 0 (no pre-selection): %q", n, view)
	}
	lower := strings.ToLower(view)
	if strings.Contains(lower, "tokens:") || strings.Contains(lower, "turns:") {
		t.Fatalf("View() still renders a metrics header: %q", view)
	}
}

// TestModelCursorStartsNoChoice (round-037 T002; supersedes round-016 T008/T017,
// architect D4): with items present, NO suggestion is the current choice at rest.
func TestModelCursorStartsNoChoice(t *testing.T) {
	m := New(context.Background(), strings.NewReader(""), &strings.Builder{}, &fakeSource{items: []string{"a", "b"}})
	if got := m.sug.selected(); got != "" {
		t.Fatalf("cursor default = %q, want no selection (empty)", got)
	}
	if n := strings.Count(m.View(), "> "); n != 0 {
		t.Fatalf("a fresh prompt marked %d suggestion row(s), want 0", n)
	}
}

// TestTabFromNoChoiceSelectsFirst (round-037 T002): the first Tab from the no-choice
// state selects (and inserts) the FIRST suggestion (reference arithmetic: -1 -> 0).
func TestTabFromNoChoiceSelectsFirst(t *testing.T) {
	m := New(context.Background(), strings.NewReader(""), &strings.Builder{}, &fakeSource{items: []string{"deploy to staging", "review the last commit"}})
	if m.sug.selected() != "" {
		t.Fatalf("expected no pre-selection, got %q", m.sug.selected())
	}
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if got := m.sug.selected(); got != "deploy to staging" {
		t.Fatalf("first Tab selected %q, want the first suggestion %q", got, "deploy to staging")
	}
	if got := m.ed.value(); got != "deploy to staging" {
		t.Fatalf("first Tab insert = %q, want %q", got, "deploy to staging")
	}
}

// TestCycleArithmetic (round-037 T004 / review F-4): freezes the reference's exact
// `(cursor+delta+len)%len` arithmetic for the no-choice sentinel and both directions,
// including the deliberate `Shift+Tab`-from-no-choice landing on `len-2` (NOT the last
// row) — the round's most surprising user-visible behaviour — and the `len == 1` case
// (Go's `%` truncates toward zero, so any integer mod 1 is 0 → the sole item IS
// selectable by both keys, not "no selection"). FR-003's wrap + `Shift+Tab` clauses.
func TestCycleArithmetic(t *testing.T) {
	cases := []struct {
		name        string
		n, cursor   int
		delta, want int
	}{
		{"len1 Tab", 1, noChoice, +1, 0},
		{"len1 ShiftTab", 1, noChoice, -1, 0},
		{"len2 Tab", 2, noChoice, +1, 0},
		{"len2 ShiftTab", 2, noChoice, -1, 0},
		{"len3 Tab", 3, noChoice, +1, 0},
		{"len3 ShiftTab", 3, noChoice, -1, 1},
		{"len3 wrap back from first", 3, 0, -1, 2},
		{"len3 wrap forward from last", 3, 2, +1, 0},
	}
	for _, tc := range cases {
		items := make([]string, tc.n)
		for i := range items {
			items[i] = fmt.Sprintf("i%d", i)
		}
		s := suggester{items: items, cursor: tc.cursor}
		s.cycle(tc.delta)
		if s.cursor != tc.want {
			t.Errorf("%s: cycle(%d) from cursor %d with len %d = %d, want %d", tc.name, tc.delta, tc.cursor, tc.n, s.cursor, tc.want)
		}
		if tc.cursor == noChoice && s.selected() == "" {
			t.Errorf("%s: a key from no-choice left no selection (cursor %d, len %d)", tc.name, s.cursor, tc.n)
		}
	}
}

// TestModelResizeReflowsWidth (round-016 T012/T017): a WindowSizeMsg sets the
// editor width, narrowing the frame.
func TestModelResizeReflowsWidth(t *testing.T) {
	m := New(context.Background(), strings.NewReader(""), &strings.Builder{}, &fakeSource{items: []string{"a"}})
	before := maxLineLen(m.View())
	_, _ = m.Update(tea.WindowSizeMsg{Width: 40})
	after := maxLineLen(m.View())
	if after >= before {
		t.Fatalf("WindowSizeMsg did not narrow the frame: before=%d after=%d", before, after)
	}
}

// TestModelTabInsertsSuggestion (round-016 T017, FR-007; round 037): Tab inserts the
// current choice — whole-line replace, and last-token replace for a multi-word
// line with a single-token suggestion (F2 unit pin). Round 037: the first Tab now
// selects the first suggestion from the no-choice state.
func TestModelTabInsertsSuggestion(t *testing.T) {
	whole := New(context.Background(), strings.NewReader(""), &strings.Builder{}, &fakeSource{items: []string{"deploy to staging"}})
	whole.ed.setValue("deploy")
	_, _ = whole.Update(tea.KeyMsg{Type: tea.KeyTab})
	if got := whole.ed.value(); got != "deploy to staging" {
		t.Fatalf("Tab whole-line insert = %q, want %q", got, "deploy to staging")
	}

	last := New(context.Background(), strings.NewReader(""), &strings.Builder{}, &fakeSource{items: []string{"staging"}})
	last.ed.setValue("deploy ")
	_, _ = last.Update(tea.KeyMsg{Type: tea.KeyTab})
	if got := last.ed.value(); got != "deploy staging" {
		t.Fatalf("Tab last-token insert = %q, want %q", got, "deploy staging")
	}
}

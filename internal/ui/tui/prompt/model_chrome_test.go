package prompt

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// fakeSource is a canned suggestion source that also records the fetch context.
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

// TestModelRendersReferenceChrome (round-016 T017): the model renders the framed
// editor + styled suggestion list, and NO session metrics header (strict parity).
func TestModelRendersReferenceChrome(t *testing.T) {
	m := New(strings.NewReader(""), &strings.Builder{}, &fakeSource{items: []string{"deploy to staging with version 016"}})
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
	if !strings.Contains(view, "> ") {
		t.Fatalf("View() missing the selection cursor: %q", view)
	}
	lower := strings.ToLower(view)
	if strings.Contains(lower, "tokens:") || strings.Contains(lower, "turns:") {
		t.Fatalf("View() still renders a metrics header: %q", view)
	}
}

// TestModelCursorDefaultsToFirstItem (round-016 T008/T017, architect D4): with
// items present, exactly the first is the current choice.
func TestModelCursorDefaultsToFirstItem(t *testing.T) {
	m := New(strings.NewReader(""), &strings.Builder{}, &fakeSource{items: []string{"a", "b"}})
	if got := m.sug.selected(); got != "a" {
		t.Fatalf("cursor default = %q, want the first item %q", got, "a")
	}
}

// TestModelResizeReflowsWidth (round-016 T012/T017): a WindowSizeMsg sets the
// editor width, narrowing the frame.
func TestModelResizeReflowsWidth(t *testing.T) {
	m := New(strings.NewReader(""), &strings.Builder{}, &fakeSource{items: []string{"a"}})
	before := maxLineLen(m.View())
	_, _ = m.Update(tea.WindowSizeMsg{Width: 40})
	after := maxLineLen(m.View())
	if after >= before {
		t.Fatalf("WindowSizeMsg did not narrow the frame: before=%d after=%d", before, after)
	}
}

// TestModelTabInsertsSuggestion (round-016 T017, FR-007): Tab inserts the
// current choice — whole-line replace, and last-token replace for a multi-word
// line with a single-token suggestion (F2 unit pin).
func TestModelTabInsertsSuggestion(t *testing.T) {
	whole := New(strings.NewReader(""), &strings.Builder{}, &fakeSource{items: []string{"deploy to staging"}})
	whole.ed.setValue("deploy")
	_, _ = whole.Update(tea.KeyMsg{Type: tea.KeyTab})
	if got := whole.ed.value(); got != "deploy to staging" {
		t.Fatalf("Tab whole-line insert = %q, want %q", got, "deploy to staging")
	}

	last := New(strings.NewReader(""), &strings.Builder{}, &fakeSource{items: []string{"staging"}})
	last.ed.setValue("deploy ")
	_, _ = last.Update(tea.KeyMsg{Type: tea.KeyTab})
	if got := last.ed.value(); got != "deploy staging" {
		t.Fatalf("Tab last-token insert = %q, want %q", got, "deploy staging")
	}
}

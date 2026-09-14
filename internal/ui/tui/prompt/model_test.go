package prompt

import (
	"strings"
	"testing"
)

// fakeSource is a canned suggestion source.
type fakeSource struct{ items []string }

func (f fakeSource) Suggest(string) []string { return f.items }

// TestModelShowsSuggestionsAndDashboard (round-015 T029): the model View renders
// the seeded suggestion and the dashboard (active provider), driven with injected
// I/O (no pty).
func TestModelShowsSuggestionsAndDashboard(t *testing.T) {
	var out strings.Builder
	m := New(strings.NewReader(""), &out, fakeSource{items: []string{"deploy to staging"}}, Dashboard{Provider: "test-model", Tokens: 5, Budget: 100, Turns: 1})
	_ = m.Init()
	view := m.View()
	if !strings.Contains(view, "deploy to staging") {
		t.Fatalf("View() = %q, want it to show the suggestion %q", view, "deploy to staging")
	}
	if !strings.Contains(view, "test-model") {
		t.Fatalf("View() = %q, want it to show the active provider %q", view, "test-model")
	}
}

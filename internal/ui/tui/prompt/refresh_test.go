package prompt

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// ctxSource records the most recent fetch context so a test can observe the
// cancel-on-keystroke behaviour (round-016 architect D1).
type ctxSource struct{ last context.Context }

func (c *ctxSource) Suggest(ctx context.Context, _ string) []string {
	c.last = ctx
	return nil
}

// TestRefreshCancelsSupersededFetch (round-016 T018, architect D1): a new
// keystroke cancels the in-flight fetch context and installs a fresh one.
func TestRefreshCancelsSupersededFetch(t *testing.T) {
	cs := &ctxSource{}
	m := New(context.Background(), strings.NewReader(""), &strings.Builder{}, cs)
	first := m.ctx
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if first.Err() == nil {
		t.Fatalf("the previous fetch context was not cancelled on a new keystroke")
	}
	if m.ctx == first {
		t.Fatalf("the model reused the cancelled fetch context")
	}
	if m.ctx.Err() != nil {
		t.Fatalf("the fresh fetch context is already cancelled")
	}
}

// TestRefreshDropsOverlongSuggestions (round-016 T014/T018, FR-006): an entry
// spanning more than three lines is not offered.
func TestRefreshDropsOverlongSuggestions(t *testing.T) {
	src := &fakeSource{items: []string{"short", "l1\nl2\nl3\nl4", "ok"}}
	m := New(context.Background(), strings.NewReader(""), &strings.Builder{}, src)
	for _, it := range m.sug.items {
		if strings.Count(strings.TrimSpace(it), "\n") >= maxSuggestionLines {
			t.Fatalf("an over-long suggestion was offered: %q", it)
		}
	}
	if len(m.sug.items) != 2 {
		t.Fatalf("expected 2 kept suggestions, got %d: %v", len(m.sug.items), m.sug.items)
	}
}

// TestRefreshUsesDebounce (round-016 T018, FR-005): the refresh is debounced, so
// the model carries the debounce duration and only recomputes on the debounce
// message for the current value.
func TestRefreshUsesDebounce(t *testing.T) {
	if DefaultDebounceDuration <= 0 {
		t.Fatalf("the debounce duration must be positive")
	}
	src := &fakeSource{items: []string{"seed"}}
	m := New(context.Background(), strings.NewReader(""), &strings.Builder{}, src)
	// A debounce message for a stale value must not recompute.
	src.items = []string{"changed"}
	_, _ = m.Update(debounceMsg{value: "stale"})
	if m.sug.selected() != "seed" {
		t.Fatalf("a stale debounce message recomputed the list")
	}
	// A debounce message for the current value recomputes.
	m.ed.setValue("seed")
	_, _ = m.Update(debounceMsg{value: m.ed.value()})
	if m.sug.selected() != "changed" {
		t.Fatalf("the current-value debounce message did not recompute the list")
	}
}

// TestModelDestroyCancelsFetch (round-016 architect TD1): Destroy cancels the
// in-flight fetch (it is the exit-time cancel invoked by Run).
func TestModelDestroyCancelsFetch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m := New(ctx, strings.NewReader(""), &strings.Builder{}, &ctxSource{})
	m.Destroy()
	if m.ctx.Err() == nil {
		t.Fatalf("Destroy() did not cancel the in-flight fetch context")
	}
}

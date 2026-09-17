package prompt

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestModelViewClearsOnSubmit (round-023 T007): before submit the frame is
// rendered; after Ctrl+S the model reports submitted and View() is empty (the
// editor frame is cleared — the authoritative teardown witness; the E2E step only
// sees the accumulated stream, so this is the exact pin).
func TestModelViewClearsOnSubmit(t *testing.T) {
	m := New(context.Background(), strings.NewReader(""), &strings.Builder{}, &fakeSource{items: []string{"hi"}})
	m.ed.setValue("hi")
	if m.View() == "" {
		t.Fatalf("View() was empty before submit")
	}
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if !m.WasSubmitted() {
		t.Fatalf("Ctrl+S did not submit")
	}
	if got := m.View(); got != "" {
		t.Fatalf("View() after submit = %q, want empty (cleared)", got)
	}
}

// TestModelViewClearsOnAbort (round-023 T007): aborting (Esc / Ctrl+C) also
// clears the frame.
func TestModelViewClearsOnAbort(t *testing.T) {
	for _, k := range []tea.KeyType{tea.KeyEsc, tea.KeyCtrlC} {
		m := New(context.Background(), strings.NewReader(""), &strings.Builder{}, &fakeSource{items: []string{"hi"}})
		if m.View() == "" {
			t.Fatalf("View() was empty before abort")
		}
		_, _ = m.Update(tea.KeyMsg{Type: k})
		if !m.WasAborted() {
			t.Fatalf("key %v did not abort", k)
		}
		if got := m.View(); got != "" {
			t.Fatalf("View() after abort = %q, want empty (cleared)", got)
		}
	}
}

// TestModelEmptySubmitKeepsFrame (round-023 FR-004; tightened by round 038,
// PR #79 review RF-3): an empty Ctrl+S does NOT submit AND returns NO command —
// the frame stays rendered and the prompt keeps running (the operator keeps
// editing). The `cmd == nil` assertion is the observable that carries the
// control-flow half of the invariant: the round-023 state-only assertions held
// even while issue #76 was live (the pre-fix code set `submitted=false` and then
// quit unconditionally). The Ctrl+S/Alt+Enter/whitespace/non-empty/abort matrix
// is owned by `model_submit_test.go` (round 038); this pin keeps the
// frame-not-cleared half.
func TestModelEmptySubmitKeepsFrame(t *testing.T) {
	m := New(context.Background(), strings.NewReader(""), &strings.Builder{}, &fakeSource{items: []string{"hi"}})
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if m.WasSubmitted() {
		t.Fatalf("an empty Ctrl+S must not submit")
	}
	if cmd != nil {
		t.Fatalf("an empty Ctrl+S returned a command; want nil (the prompt must keep running)")
	}
	if m.View() == "" {
		t.Fatalf("the frame was cleared on an empty submit")
	}
}

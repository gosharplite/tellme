package prompt

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Round-038 unit pins (issue #76): submitting an EMPTY editor is a no-op (the
// model returns NO command, so the prompt stays open), while a non-empty submit
// and an abort still quit.

func newSubmitTestModel() *Model {
	return New(context.Background(), strings.NewReader(""), &strings.Builder{}, &fakeSource{items: []string{"hi"}})
}

func TestModelEmptySubmitDoesNotQuit(t *testing.T) {
	for _, k := range []tea.KeyMsg{
		{Type: tea.KeyCtrlS},
		{Type: tea.KeyEnter, Alt: true},
	} {
		m := newSubmitTestModel()
		_, cmd := m.Update(k)
		if cmd != nil {
			t.Fatalf("key %v: an empty submit returned a command; want nil (the prompt stays open)", k)
		}
		if m.WasSubmitted() {
			t.Fatalf("key %v: an empty submit marked the model submitted", k)
		}
		if m.WasAborted() {
			t.Fatalf("key %v: an empty submit marked the model aborted", k)
		}
		if m.View() == "" {
			t.Fatalf("key %v: the frame was cleared on an empty submit", k)
		}
	}
}

func TestModelWhitespaceOnlySubmitDoesNotQuit(t *testing.T) {
	m := newSubmitTestModel()
	m.ed.setValue("   ")
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if cmd != nil || m.WasSubmitted() {
		t.Fatalf("a whitespace-only submit must be a no-op (cmd=%v submitted=%v)", cmd, m.WasSubmitted())
	}
}

func TestModelNonEmptySubmitQuits(t *testing.T) {
	m := newSubmitTestModel()
	m.ed.setValue("hi")
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if !m.WasSubmitted() {
		t.Fatalf("a non-empty Ctrl+S did not submit")
	}
	if cmd == nil {
		t.Fatalf("a non-empty submit returned no command; want a quit command")
	}
}

func TestModelAbortStillQuits(t *testing.T) {
	for _, k := range []tea.KeyType{tea.KeyEsc, tea.KeyCtrlC} {
		m := newSubmitTestModel()
		_, cmd := m.Update(tea.KeyMsg{Type: k})
		if !m.WasAborted() || cmd == nil {
			t.Fatalf("abort key %v did not quit (aborted=%v cmd=%v)", k, m.WasAborted(), cmd)
		}
	}
}
